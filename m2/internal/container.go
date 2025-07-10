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

	Done chan ContainerResult

	programs []*program.Program

	stdin  <-chan string
	stdout chan<- string
	stderr chan<- string

	user string
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
	}
	for _, arg := range opts {
		arg(p)
	}
	return p
}

var ErrNoProgram = fmt.Errorf("no programs registered")

func (c *Container) Run() (id string, err error) {
	if len(c.programs) == 0 {
		return "", ErrNoProgram
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

	var apiClient *client.Client
	apiClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	ctx := context.Background()

	resp, err := apiClient.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}
	id = resp.ID
	// Attach then start:
	// https://stackoverflow.com/questions/65283411/docker-attach-vs-docker-start-ai-for-a-running-container
	var conn types.HijackedResponse
	conn, err = apiClient.ContainerAttach(ctx, id, container.AttachOptions{
		Stream: true,
		Stdin:  attachStdin,
		Stdout: attachStdout,
		Stderr: attachStderr,
	})
	if err != nil {
		return id, err
	}

	var wg sync.WaitGroup

	if attachStdout || attachStderr {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Close output channels when container outputs close
			if attachStdout {
				defer close(c.stdout)
			}
			if attachStderr {
				defer close(c.stderr)
			}

			stdcopy.StdCopy(
				NewChannelWriter(c.stdout),
				NewChannelWriter(c.stderr),
				conn.Reader)
		}()
	}
	if attachStdin {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer conn.CloseWrite() // Close when the stdin channel closes

			for input := range c.stdin {
				if _, err := conn.Conn.Write([]byte(input)); err != nil {
					return
				}
			}
		}()
	} else {
		conn.CloseWrite()
	}

	statusChan, errChan := apiClient.ContainerWait(ctx, id, container.WaitConditionNextExit)

	go func() {
		wg.Wait() // Wait for stdin, stdout, and stderr to close
		conn.Close()

		select {
		case status := <-statusChan:
			c.Done <- ContainerResult{
				ID:     id,
				Status: int(status.StatusCode),
				Err:    nil,
			}
		case err := <-errChan:
			c.Done <- ContainerResult{
				ID:     id,
				Status: -1,
				Err:    err,
			}
		}

		close(c.Done)

		apiClient.ContainerRemove(ctx, id, container.RemoveOptions{})
		apiClient.Close()
	}()

	if err := apiClient.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return id, err
	}

	return id, nil
}

func (c *Container) BuildProgramCmd() (cmdTokens []string, mounts []mount.Mount) {
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
