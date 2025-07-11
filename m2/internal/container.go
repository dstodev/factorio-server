package internal

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/google/uuid"

	"manage2/internal/program"
	"manage2/internal/stream"
)

type Container struct {
	Image string
	ID    string

	Done chan ContainerResult

	streams *stream.ContainerStream

	user string

	hostToGuestMap map[string]string // Maps host paths to guest paths

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
		Image:          image,
		Done:           make(chan ContainerResult, 1),
		streams:        stream.NewContainerStream(),
		hostToGuestMap: make(map[string]string),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

var ErrNoProgram = fmt.Errorf("no programs registered")

// Run invokes a command in the container, mounting file arguments. The first
// program is expected to parse all arguments up to the "--" separator argument.
// The first following string is the next program's path, and the rest are its
// arguments. e.g.:
//
//	first, err := program.FromFile(
//	    "/my/script1.sh",
//	    WithStringArgs("hello,"))
//	second, err := program.FromFile(
//	    "/sub/script2.sh",
//	    WithStringArgs("world!"),
//	    WithFileArgs("/host-path/file.txt"))
//
//	c := NewContainer("my-image:latest")
//	err = c.Run(first, second)
//
// Here, /my/script1.sh is called inside the container like:
//
//	/mp/script1.sh "hello," "--" "/mp/script2.sh" "world!" "/mp/file.txt" "--"
//
// (/mp/ refers to a unique mount path for each file)
//
// It is assumed script1.sh will parse its own arguments, run, then start
// script2.sh with remaining arguments.
func (c *Container) Run(programs ...*program.Program) error {
	if len(programs) == 0 {
		return ErrNoProgram
	}

	attachStdin := c.streams.Stdin != nil
	attachStdout := c.streams.Stdout != nil
	attachStderr := c.streams.Stderr != nil

	var err error

	c.client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	c.ctx = context.Background()

	cmdTokens, mounts := c.BuildProgramCmd(programs...)

	// Mounts are also recorded by WithMounts() so Exec() can use them
	for _, mount := range mounts {
		c.hostToGuestMap[mount.Source] = mount.Target
	}

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

	resp, err := c.client.ContainerCreate(c.ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return err
	}
	c.ID = resp.ID

	// Attach then start:
	// https://github.com/docker/cli/blob/578ccf607d24abc5270e9a4cbd5ba9b5355b042f/cli/command/container/start.go#L113
	// https://github.com/docker/cli/blob/578ccf607d24abc5270e9a4cbd5ba9b5355b042f/cli/command/container/start.go#L146

	ctrSock, err := c.client.ContainerAttach(c.ctx, c.ID, container.AttachOptions{
		Stream: true,
		Stdin:  attachStdin,
		Stdout: attachStdout,
		Stderr: attachStderr,
	})
	if err != nil {
		return err
	}

	c.streams.StartStreaming(&ctrSock)

	c.startExitHandler()

	if err := c.client.ContainerStart(c.ctx, c.ID, container.StartOptions{}); err != nil {
		return err
	}

	return nil
}

// BuildProgramCmd takes registered programs and constructs a command line to
// run in the container. Guest paths are generated for file arguments referenced
// by the programs, and mounts are created to bind the host paths to the guest
// paths. Returns the command and mounts to use when starting the container.
func (c *Container) BuildProgramCmd(programs ...*program.Program) (
	cmdTokens []string,
	mounts []mount.Mount,
) {
	toGuestMutator := func(p *program.Program, index int, arg string) string {
		if p.ArgIsFile(index) {
			hostPath := arg
			var guestPath string
			if cachedPath, ok := c.hostToGuestMap[hostPath]; ok {
				// If the program already has a mapping for this path, use it
				guestPath = cachedPath
			} else {
				guestPath = toGuestPath(hostPath)
				c.hostToGuestMap[hostPath] = guestPath
			}
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: hostPath,
				Target: guestPath,
			})
			return guestPath
		}
		return arg
	}
	for _, p := range programs {
		cmdTokens = append(cmdTokens, p.AsTokens(toGuestMutator)...)
	}
	return cmdTokens, mounts
}

func toGuestPath(hostPath string) string {
	guestID := uuid.New().String()
	guestBasename := filepath.Base(hostPath)
	guestPath := fmt.Sprintf("/tmp/m2/%s/%s", guestID, guestBasename)
	return guestPath
}

// startExitHandler waits for the container to exit, then sends the result to
// the Done channel. It then removes the container.
func (c *Container) startExitHandler() {
	statusChan, errChan := c.client.ContainerWait(c.ctx, c.ID, container.WaitConditionNextExit)

	go func() {
		c.streams.WaitUntilClosed()

		select {
		case status := <-statusChan:
			c.Done <- ContainerResult{
				ID:     c.ID,
				Status: int(status.StatusCode),
				Err:    nil,
			}
		case err := <-errChan:
			c.Done <- ContainerResult{
				ID:     c.ID,
				Status: -1,
				Err:    err,
			}
		}

		close(c.Done)

		c.client.ContainerRemove(c.ctx, c.ID, container.RemoveOptions{})
		c.client.Close()
	}()
}

var ErrNotRunning = fmt.Errorf("container is not running")
var ErrNotMounted = fmt.Errorf("container has not mounted all file arguments")

// Exec runs a program in a running container.
//
// File arguments added to a program using WithFileArgs() require special
// handling: Exec() will return an error if the program has any file arguments
// that were not mounted when the container was create. Use ContainerOption
// WithMounts() to do so. Docker does not support mounting additional files
// after the container has started.
func (c *Container) Exec(p *program.Program, opts ...stream.Option) <-chan ContainerResult {
	resultChan := make(chan ContainerResult, 1)
	result := ContainerResult{
		ID:     "",
		Status: -1,
		Err:    nil,
	}
	sendErr := func(err error) <-chan ContainerResult {
		result.Err = err
		resultChan <- result
		return resultChan
	}

	if c.ID == "" {
		return sendErr(ErrNotRunning)
	}

	cmdTokens, mounts := c.BuildProgramCmd(p)

	for _, mount := range mounts {
		if _, ok := c.hostToGuestMap[mount.Source]; !ok {
			return sendErr(ErrNotMounted)
		}
	}

	cs := stream.NewContainerStream(opts...)
	attachStdin := cs.Stdin != nil
	attachStdout := cs.Stdout != nil
	attachStderr := cs.Stderr != nil

	resp, err := c.client.ContainerExecCreate(c.ctx, c.ID, container.ExecOptions{
		Cmd:          cmdTokens,
		Tty:          false,
		User:         c.user,
		AttachStdin:  attachStdin,
		AttachStdout: attachStdout,
		AttachStderr: attachStderr,
	})
	if err != nil {
		return sendErr(err)
	}
	result.ID = resp.ID

	ctrSock, err := c.client.ContainerExecAttach(c.ctx, resp.ID, container.ExecAttachOptions{
		Tty: false,
	})
	if err != nil {
		return sendErr(err)
	}

	cs.StartStreaming(&ctrSock)

	go func() {
		cs.WaitUntilClosed()

		status, err := c.client.ContainerExecInspect(c.ctx, resp.ID)
		if err != nil {
			sendErr(err)
			return
		}
		result.Status = status.ExitCode
		resultChan <- result
		close(resultChan)
	}()

	err = c.client.ContainerExecStart(c.ctx, resp.ID, container.ExecStartOptions{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return sendErr(err)
	}

	return resultChan
}
