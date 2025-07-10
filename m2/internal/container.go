package internal

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"

	"manage2/internal/program"
)

type Container struct {
	Image string

	Id   string
	Done chan ContainerResult

	programs []*program.Program

	stdin  <-chan string
	stdout chan<- string
	stderr chan<- string

	user string

	wg *sync.WaitGroup

	ctx    context.Context
	client *client.Client
}

type ContainerResult struct {
	ID     string
	Status int
	Err    error
}

func NewContainer(image string, opts ...ContainerOption) *Container {
	p := &Container{
		Image: image,
		Done:  make(chan ContainerResult, 1),
		wg:    &sync.WaitGroup{},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

var ErrNoProgram = fmt.Errorf("no programs registered")

func (c *Container) Run() error {
	if len(c.programs) == 0 {
		return ErrNoProgram
	}

	cmdTokens, mounts := c.BuildProgramCmd()

	attachStdin := c.stdin != nil
	attachStdout := c.stdout != nil
	attachStderr := c.stderr != nil

	config := &container.Config{
		Image:        c.Image,
		Cmd:          cmdTokens,
		Tty:          false,
		User:         c.user,
		AttachStdin:  attachStdin,
		OpenStdin:    attachStdin,
		StdinOnce:    attachStdin,
		AttachStdout: attachStdout,
		AttachStderr: attachStderr,
	}

	hostConfig := &container.HostConfig{
		Mounts: mounts,
	}

	var err error

	c.client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	c.ctx = context.Background()

	resp, err := c.client.ContainerCreate(c.ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return err
	}
	c.Id = resp.ID

	// Attach then start:
	// https://github.com/docker/cli/blob/master/cli/command/container/start.go#L113
	// https://github.com/docker/cli/blob/master/cli/command/container/start.go#L146

	conn, err := c.client.ContainerAttach(c.ctx, c.Id, container.AttachOptions{
		Stream: true,
		Stdin:  attachStdin,
		Stdout: attachStdout,
		Stderr: attachStderr,
	})
	if err != nil {
		return err
	}

	if attachStdout || attachStderr {
		c.startWritingOutputToChannels(&conn, attachStdout, attachStderr)
	}

	if attachStdin {
		c.startReadingInputFromChannel(&conn)
	} else {
		conn.CloseWrite()
	}

	c.startExitHandler(&conn)

	if err := c.client.ContainerStart(c.ctx, c.Id, container.StartOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *Container) BuildProgramCmd() (
	cmdTokens []string,
	mounts []mount.Mount,
) {
	toGuestMutator := func(p *program.Program, index int, arg string) string {
		if p.ArgIsFile(index) {
			hostPath := arg
			guestId := uuid.New().String()
			guestBasename := filepath.Base(hostPath)
			guestPath := fmt.Sprintf("/tmp/m2/%s/%s", guestId, guestBasename)
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: hostPath,
				Target: guestPath,
			})
			return guestPath
		}
		return arg
	}
	for _, p := range c.programs {
		cmdTokens = append(cmdTokens, p.AsTokens(toGuestMutator)...)
	}
	return cmdTokens, mounts
}

func (c *Container) startWritingOutputToChannels(
	conn *types.HijackedResponse,
	closeStdout bool,
	closeStderr bool,
) {
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		// Close output channels when container outputs close
		if closeStdout {
			defer close(c.stdout)
		}
		if closeStderr {
			defer close(c.stderr)
		}

		stdcopy.StdCopy(
			NewChannelWriter(c.stdout),
			NewChannelWriter(c.stderr),
			conn.Reader)
	}()
}

func (c *Container) startReadingInputFromChannel(conn *types.HijackedResponse) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer conn.CloseWrite() // Close when the stdin channel closes

		for input := range c.stdin {
			if _, err := conn.Conn.Write([]byte(input)); err != nil {
				return
			}
		}
	}()
}

func (c *Container) startExitHandler(conn *types.HijackedResponse) {
	statusChan, errChan := c.client.ContainerWait(c.ctx, c.Id, container.WaitConditionNextExit)

	go func() {
		c.wg.Wait() // Wait for stdin, stdout, and stderr to close
		conn.Close()

		select {
		case status := <-statusChan:
			c.Done <- ContainerResult{
				ID:     c.Id,
				Status: int(status.StatusCode),
				Err:    nil,
			}
		case err := <-errChan:
			c.Done <- ContainerResult{
				ID:     c.Id,
				Status: -1,
				Err:    err,
			}
		}

		close(c.Done)

		c.client.ContainerRemove(c.ctx, c.Id, container.RemoveOptions{})
		c.client.Close()
	}()
}
