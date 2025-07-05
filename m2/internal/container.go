package internal

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"

	"manage2/internal/program"
)

type Container struct {
	Image string

	programs []*program.Program

	stdin  <-chan string
	stdout chan<- string
	stderr chan<- string

	user string
}

type ContainerArgument func(*Container)

func NewContainer(image string, args ...ContainerArgument) *Container {
	p := &Container{
		Image: image,
	}
	return p
}

// WithProgram adds a program to run, mounting file arguments. The first
// program is expected to parse all arguments up to the "--" separator argument.
// The first following string is the next program's path, and the rest are its
// arguments. e.g.:
//
//	NewContainer(
//	  "some-image:latest",
//	  WithProgram(NewProgram(
//	    "my/script.sh",
//	    WithStringArgs("hello,")
//	  )),
//	  WithProgram(NewProgram(
//	    "sub/script.sh",
//	    WithStringArgs("world!"),
//	    WithFileArgs("/host-path/file.txt")
//	  )),
//	)
//
// Here, my/script.sh is called inside the container like:
//
//	my/script.sh "hello," "--" "sub/script.sh" "world!" "/mount-path/file.txt" "--"
//
// and it is assumed my/script.sh will parse its arguments, run, and then start
// sub/script.sh with its arguments.
func WithProgram(program *program.Program) ContainerArgument {
	return func(p *Container) {
		p.programs = append(p.programs, program)
	}
}

func WithStdinChannel(stdin <-chan string) ContainerArgument {
	return func(p *Container) {
		p.stdin = stdin
	}
}

func WithStdoutChannel(stdout chan<- string) ContainerArgument {
	return func(p *Container) {
		p.stdout = stdout
	}
}

func WithStderrChannel(stderr chan<- string) ContainerArgument {
	return func(p *Container) {
		p.stderr = stderr
	}
}

func WithUser(user string) ContainerArgument {
	return func(p *Container) {
		p.user = user
	}
}

type mountPath struct {
	Host  string
	Guest string
}

func (p *Container) Run() error {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer apiClient.Close()

	attachStdin := p.stdin != nil
	attachStdout := p.stdout != nil
	attachStderr := p.stderr != nil

	var mountPaths []mountPath

	toGuestMutator := func(p *program.Program, index int, arg string) string {
		if p.ArgIsFile(index) {
			hostPath := arg
			guestId := uuid.New().String()
			guestBasename := filepath.Base(hostPath)
			guestPath := fmt.Sprintf("/tmp/m2/%s/%s", guestId, guestBasename)
			mountPaths = append(mountPaths, mountPath{
				Host:  hostPath,
				Guest: guestPath,
			})
			return guestPath
		}
		return arg
	}

	var cmdTokens []string

	for _, program := range p.programs {
		cmdTokens = append(cmdTokens, program.AsTokens(toGuestMutator)...)
	}

	config := &container.Config{
		Image:        p.Image,
		Cmd:          cmdTokens,
		Tty:          false,
		AttachStdin:  attachStdin,
		OpenStdin:    attachStdin,
		StdinOnce:    attachStdin,
		AttachStdout: attachStdout,
		AttachStderr: attachStderr,
	}

	var mounts []mount.Mount

	for _, mountPath := range mountPaths {
		mount := mount.Mount{
			Type:   mount.TypeBind,
			Source: mountPath.Host,
			Target: mountPath.Guest,
		}
		mounts = append(mounts, mount)
	}

	hostConfig := &container.HostConfig{
		Mounts: mounts,
	}

	ctx := context.Background()
	resp, err := apiClient.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return err
	}
	// Attach then start:
	// https://stackoverflow.com/questions/65283411/docker-attach-vs-docker-start-ai-for-a-running-container
	conn, err := apiClient.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: attachStdout || attachStderr,
		Stdin:  attachStdin,
		Stdout: attachStdout,
		Stderr: attachStderr,
	})
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	if attachStdout || attachStderr {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stdcopy.StdCopy(
				NewChannelWriter(p.stdout),
				NewChannelWriter(p.stderr),
				conn.Reader)
		}()
	}
	if attachStdin {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for input := range p.stdin {
				if _, err := conn.Conn.Write([]byte(input)); err != nil {
					return
				}
			}
		}()
	}

	go func() {
		defer conn.Close()
		wg.Wait()
	}()

	if err := apiClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	return nil
}
