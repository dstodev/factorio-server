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

	"manage2/internal/container/stream"
	"manage2/internal/program"
)

type Container struct {
	Image string
	ID    string

	HostToGuestMap map[string]string

	cs   *stream.ContainerStream
	user string

	ctx    context.Context
	client *client.Client
}

type Result struct {
	ID     string
	Status int
	Err    error
}

func New(image string, opts ...ContainerOption) *Container {
	c := &Container{
		Image:          image,
		cs:             stream.NewContainerStream(),
		HostToGuestMap: make(map[string]string),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Idle creates a container that remains open but idle until ctrClose() is
// called. This is useful to run an arbitrary number of commands with Exec().
// WithStdinChannel() options are ignored.
func Idle(image string, opts ...ContainerOption) (ctr *Container, ctrClose func() Result) {
	ctrStdin := make(chan string)
	opts = append(opts, WithStdinChannel(ctrStdin))
	ctr = New(image, opts...)
	pgm, err := program.FromSystem("cat")
	if err != nil {
		return nil, nil
	}
	ctrDone, err := ctr.Run(pgm)
	if err != nil {
		return nil, nil
	}
	ctrClose = func() Result {
		// Allow & wait for container to close
		close(ctrStdin)
		return <-ctrDone
	}
	return
}

// TODO: Find() to find a container by ID or image name, returning a Container

var ErrNotNew = fmt.Errorf("container must be new to run programs")
var ErrNoProgram = fmt.Errorf("no programs registered")

// Run invokes a command in the container, mounting file arguments. The first
// program is expected to parse all arguments up to the "--" separator argument.
// The first following string is the next program's path, and the rest are its
// arguments. e.g.:
//
//	first, err := program.FromFile(
//	    "/host/script1.sh",
//	    WithStringArgs("hello,"))
//	second, err := program.FromFile(
//	    "/sub/script2.sh",
//	    WithStringArgs("world!"),
//	    WithFileArgs("/host-path/file.txt"))
//
//	ctr := container.New("my-image:latest")
//	err = ctr.Run(first, second)
//
// Here, /host/script1.sh is called inside the container like:
//
//	/mp/script1.sh "hello," "--" "/mp/script2.sh" "world!" "/mp/file.txt" "--"
//
// (/mp/ refers to a unique mount path for each file)
//
// It is assumed script1.sh will parse its own arguments, run, then start
// script2.sh with remaining arguments.
func (c *Container) Run(programs ...*program.Program) (ctrDone <-chan Result, err error) {
	if c.client != nil {
		return nil, ErrNotNew
	}

	if len(programs) == 0 {
		return nil, ErrNoProgram
	}

	c.client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	c.ctx = context.Background()

	cmdTokens, fileMap := c.BuildProgramCmd(programs...)

	// Mounts are also recorded by WithMounts() so Exec() can use them
	for hostPath, guestPath := range fileMap {
		if _, ok := c.HostToGuestMap[hostPath]; !ok {
			c.HostToGuestMap[hostPath] = guestPath
		}
	}

	attachStdin := c.cs.Stdin != nil
	attachStdout := c.cs.Stdout != nil
	attachStderr := c.cs.Stderr != nil

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

	mounts := make([]mount.Mount, 0, len(c.HostToGuestMap))

	for hostPath, guestPath := range c.HostToGuestMap {
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
		return nil, err
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
		return nil, err
	}

	done := make(chan Result, 1)

	c.cs.StartStreaming(&ctrSock)

	c.startExitHandler(done)

	if err := c.client.ContainerStart(c.ctx, c.ID, container.StartOptions{}); err != nil {
		return nil, err
	}

	return done, nil
}

// BuildProgramCmd takes registered programs and constructs a command line to
// run in the container. Guest paths are generated for file arguments referenced
// by the programs.
//
// Returns the command to use when starting the container, and a list of new
// mappings from host paths to guest paths. Run() uses this list to mount them.
func (c *Container) BuildProgramCmd(programs ...*program.Program) (
	cmdTokens []string,
	fileMap map[string]string,
) {
	fileMap = make(map[string]string)

	toGuestMutator := func(pgm *program.Program, index int, arg string) string {
		if pgm.ArgIsFile(index) {
			hostPath := arg
			var guestPath string

			if path, ok := c.HostToGuestMap[hostPath]; ok {
				guestPath = path
			} else if path, ok := fileMap[hostPath]; ok {
				guestPath = path
			} else {
				guestPath = toGuestPath(hostPath)
				fileMap[hostPath] = guestPath
			}

			return guestPath
		}
		return arg
	}
	for _, pgm := range programs {
		cmdTokens = append(cmdTokens, pgm.AsTokens(toGuestMutator)...)
	}
	return
}

func toGuestPath(hostPath string) string {
	guestID := uuid.New().String()
	guestBasename := filepath.Base(hostPath)
	guestPath := fmt.Sprintf("/tmp/m2/%s/%s", guestID, guestBasename)
	return guestPath
}

// startExitHandler waits for the container to exit, then sends the result to
// the Done channel. It then removes the container.
func (c *Container) startExitHandler(done chan<- Result) {
	ctrStatus, ctrErr := c.client.ContainerWait(c.ctx, c.ID, container.WaitConditionNextExit)

	go func() {
		defer close(done)

		result := Result{
			ID:     c.ID,
			Status: -1,
			Err:    nil,
		}

		select {
		case status := <-ctrStatus:
			result.Status = int(status.StatusCode)
		case err := <-ctrErr:
			result.Err = err
		}

		c.cs.Close()

		c.client.ContainerRemove(c.ctx, c.ID, container.RemoveOptions{})
		c.client.Close()

		done <- result
	}()
}

var ErrNotRunning = fmt.Errorf("container is not running")
var ErrNotMounted = fmt.Errorf("container has not mounted all file arguments")

// Exec runs a program in a running container.
//
// File arguments added to a program using WithFileArgs() require special
// handling: Exec() will return ErrNotMounted if the program has any file
// arguments that were not mounted when the container was created. Use
// ContainerOption WithMounts() to do so.
func (c *Container) Exec(
	pgm *program.Program,
	opts ...stream.StreamOption,
) (
	execDone <-chan Result,
	err error,
) {
	if pgm == nil {
		return nil, ErrNoProgram
	}

	if c.ID == "" {
		return nil, ErrNotRunning
	}

	cmdTokens, fileMap := c.BuildProgramCmd(pgm)

	// If any mappings are new here, they are not yet mounted. There is no way
	// to mount additional files after the container has started.
	if len(fileMap) > 0 {
		return nil, ErrNotMounted
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
		return nil, err
	}
	result := Result{
		ID:     resp.ID,
		Status: -1,
		Err:    nil,
	}

	ctrSock, err := c.client.ContainerExecAttach(c.ctx, result.ID, container.ExecAttachOptions{
		Tty: false,
	})
	if err != nil {
		return nil, err
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

	execResult := make(chan Result, 1)

	go func() {
		defer close(execResult)

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

		cs.Close()

		status, err := c.client.ContainerExecInspect(c.ctx, resp.ID)
		result.Err = err
		result.Status = status.ExitCode
		execResult <- result
	}()

	return execResult, nil
}
