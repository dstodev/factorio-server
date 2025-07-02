package internal

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
)

type ProgramArgument func(*Program)

type Mount struct {
	Host  string
	Guest string
}

type Program struct {
	Path  string
	Image string

	args   []string
	mounts []Mount

	stdin  <-chan string
	stdout chan<- string
	stderr chan<- string

	user string
}

func NewProgram(path string, image string, args ...ProgramArgument) *Program {
	p := &Program{
		Path:  path,
		Image: image,
		args:  []string{path},
	}
	WithMounts(path)(p)
	for _, arg := range args {
		arg(p)
	}
	return p
}

func WithArgs(args ...string) ProgramArgument {
	return func(p *Program) {
		p.args = append(p.args, args...)
	}
}

// WithMounts adds file mounts for use when running the program in its container.
//
// Additionally adds the mounts to the program's arguments by mounted filepath.
func WithMounts(mounts ...string) ProgramArgument {
	m := toMounts(mounts)

	return func(p *Program) {
		p.mounts = append(p.mounts, m...)
		p.args = append(p.args, toGuestPaths(m)...)
	}
}

func toMounts(mounts []string) []Mount {
	result := make([]Mount, len(mounts))
	for i, hostPath := range mounts {
		id := uuid.New().String()
		basename := filepath.Base(hostPath)
		guestPath := fmt.Sprintf("/tmp/m2/%s/%s", id, basename)

		result[i] = Mount{
			Host:  hostPath,
			Guest: guestPath,
		}
	}
	return result
}

func toGuestPaths(mounts []Mount) []string {
	result := make([]string, len(mounts))
	for i, mount := range mounts {
		result[i] = mount.Guest
	}
	return result
}

// WithProgram adds a sub-program's path as an argument, and mounts its mounts.
func WithProgram(subProgram *Program) ProgramArgument {
	return func(p *Program) {
		p.args = append(p.args, subProgram.Path)
		p.mounts = append(p.mounts, subProgram.mounts...)
	}
}

func WithStdinChannel(stdin <-chan string) ProgramArgument {
	return func(p *Program) {
		p.stdin = stdin
	}
}

func WithStdoutChannel(stdout chan<- string) ProgramArgument {
	return func(p *Program) {
		p.stdout = stdout
	}
}

func WithStderrChannel(stderr chan<- string) ProgramArgument {
	return func(p *Program) {
		p.stderr = stderr
	}
}

func WithUser(user string) ProgramArgument {
	return func(p *Program) {
		p.user = user
	}
}

func (p *Program) Run() error {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer apiClient.Close()

	attachStdin := p.stdin != nil
	attachStdout := p.stdout != nil
	attachStderr := p.stderr != nil

	config := &container.Config{
		Image:        p.Image,
		Cmd:          p.args,
		Tty:          false,
		AttachStdin:  attachStdin,
		OpenStdin:    attachStdin,
		StdinOnce:    attachStdin,
		AttachStdout: attachStdout,
		AttachStderr: attachStderr,
	}

	mounts := make([]mount.Mount, len(p.mounts))
	for i, m := range p.mounts {
		mounts[i] = mount.Mount{
			Type:   mount.TypeBind,
			Source: m.Host,
			Target: m.Guest,
		}
	}

	hostConfig := &container.HostConfig{
		Mounts: mounts,
	}

	ctx := context.Background()
	resp, err := apiClient.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return err
	}
	conn, err := apiClient.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: attachStdout || attachStderr,
		Stdin:  attachStdin,
		Stdout: attachStdout,
		Stderr: attachStderr,
	})
	if err != nil {
		return err
	}
	defer conn.Close()
	if attachStdin {
		go func() {
			for input := range p.stdin {
				if _, err := conn.Conn.Write([]byte(input)); err != nil {
					return
				}
			}
		}()
	}
	if attachStdout || attachStderr {
		go stdcopy.StdCopy(p.stdout, p.stderr, conn.Reader)
	}
	return nil
}
