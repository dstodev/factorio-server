package container

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/google/uuid"

	"manage2/internal/program"
	"manage2/internal/stream"
)

type Container struct {
	Image string
	ID    string

	Done chan Result

	streams *stream.ContainerStream

	user string

	hostToGuestMap map[string]string

	ctx    context.Context
	client *client.Client
}

type Result struct {
	ID     string
	Status int
	Err    error
}

func New(image string, opts ...Option) *Container {
	c := &Container{
		Image:          image,
		Done:           make(chan Result, 1),
		streams:        stream.NewContainerStream(),
		hostToGuestMap: make(map[string]string),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
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
//	c := container.New("my-image:latest")
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

	cmdTokens, fileMap := c.BuildProgramCmd(programs...)

	// Mounts are also recorded by WithMounts() so Exec() can use them
	for hostPath, guestPath := range fileMap {
		if _, ok := c.hostToGuestMap[hostPath]; !ok {
			c.hostToGuestMap[hostPath] = guestPath
		}
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

	mounts := make([]mount.Mount, 0, len(c.hostToGuestMap))

	for hostPath, guestPath := range c.hostToGuestMap {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: hostPath,
			Target: guestPath,
		})
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
	fileMap map[string]string,
) {
	fileMap = make(map[string]string)
	toGuestMutator := func(pgm *program.Program, index int, arg string) string {
		if pgm.ArgIsFile(index) {
			hostPath := arg
			guestPath := c.resolveGuestPath(hostPath)
			fileMap[hostPath] = guestPath
			return guestPath
		}
		return arg
	}
	for _, pgm := range programs {
		cmdTokens = append(cmdTokens, pgm.AsTokens(toGuestMutator)...)
	}
	return cmdTokens, fileMap
}

// resolveGuestPath returns the guest path for a given host path, creating a new
// guest path if it does not already exist. New paths are not added to the
// hostToGuestMap automatically. Instead, Run() records new paths before
// starting the container. This enables Exec() to accurately verify all required
// files were mounted by ensuring all files referenced by a program are mapped
// to a guest in hostToGuestMap.
func (c *Container) resolveGuestPath(hostPath string) (guestPath string) {
	if path, ok := c.hostToGuestMap[hostPath]; ok {
		guestPath = path
	} else {
		guestPath = toGuestPath(hostPath)
	}
	return guestPath
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
	ctrStatus, ctrErr := c.client.ContainerWait(c.ctx, c.ID, container.WaitConditionNextExit)

	go func() {
		defer close(c.Done)
		c.streams.WaitUntilClosed()

		select {
		case status := <-ctrStatus:
			c.Done <- Result{
				ID:     c.ID,
				Status: int(status.StatusCode),
				Err:    nil,
			}
		case err := <-ctrErr:
			c.Done <- Result{
				ID:     c.ID,
				Status: -1,
				Err:    err,
			}
		}

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
func (c *Container) Exec(pgm *program.Program, opts ...stream.Option) <-chan Result {
	execResult := make(chan Result, 1)
	result := Result{
		ID:     "",
		Status: -1,
		Err:    nil,
	}
	sendErr := func(err error) <-chan Result {
		result.Err = err
		execResult <- result
		return execResult
	}

	if pgm == nil {
		return sendErr(ErrNoProgram)
	}

	if c.ID == "" {
		return sendErr(ErrNotRunning)
	}

	cmdTokens, fileMap := c.BuildProgramCmd(pgm)

	for hostPath := range fileMap {
		// c.hostToGuestMap is only written by WithMounts() and by Run() for
		// initial mounts, so verify all referenced files are mounted.
		// (Docker does not support new mounts after a container has started.)
		if _, ok := c.hostToGuestMap[hostPath]; !ok {
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

	ctrSock, err := c.client.ContainerExecAttach(c.ctx, result.ID, container.ExecAttachOptions{
		Tty: false,
	})
	if err != nil {
		return sendErr(err)
	}
	cs.StartStreaming(&ctrSock)

	// Start listening for "exec_die" events from this container
	since := time.Now().Format(time.RFC3339)
	eventStreamCtx, eventStreamCancel := context.WithCancel(c.ctx)
	msgs, errs := c.client.Events(eventStreamCtx, events.ListOptions{
		Since: since,
		Filters: filters.NewArgs(
			// https://docs.docker.com/reference/cli/docker/system/events/#filter
			filters.Arg("type", "container"),
			filters.Arg("event", "exec_die"),
			filters.Arg("container", c.ID),
		),
	})

	go func() {
		defer close(execResult)
		cs.WaitUntilClosed()

		// Wait for the exec'd program to finish by waiting for its exec_die event.
		for {
			done := false
			select {
			case msg := <-msgs:
				execID := msg.Actor.Attributes["execID"]
				if execID == result.ID {
					done = true
				}
			case <-errs:
				done = true
			}
			if done {
				eventStreamCancel()
				break
			}
		}

		status, err := c.client.ContainerExecInspect(c.ctx, resp.ID)
		if err != nil {
			sendErr(err)
			return
		}
		result.Status = status.ExitCode
		execResult <- result
	}()

	return execResult
}
