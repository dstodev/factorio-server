package internal

import "manage2/internal/program"

type ContainerOption func(c *Container)

// WithProgram adds a program to run, mounting file arguments. The first program
// is expected to parse all arguments up to the "--" separator argument. The
// first following string is the next program's path, and the rest are its
// arguments. e.g.:
//
//	NewContainer(
//	  "some-image:latest",
//	  WithProgram(FromFile(
//	    "/my/script1.sh",
//	    WithStringArgs("hello,")
//	  )),
//	  WithProgram(FromFile(
//	    "/sub/script2.sh",
//	    WithStringArgs("world!"),
//	    WithFileArgs("/host-path/file.txt")
//	  )),
//	)
//
// (code is abbreviated; excludes error handling)
//
// Here, /my/script1.sh is called inside the container like:
//
//	/mp/script1.sh "hello," "--" "/mp/script2.sh" "world!" "/mp/file.txt" "--"
//
// (/mp/ refers to the mount path inside the container, which is unique for
// each file)
//
// and it is assumed script1.sh will parse its own arguments, run, and then
// start script2.sh with remaining arguments.
func WithProgram(program *program.Program) ContainerOption {
	return func(c *Container) {
		c.programs = append(c.programs, program)
	}
}

func WithStdinChannel(stdin <-chan string) ContainerOption {
	return func(c *Container) {
		c.stdin = stdin
	}
}

func WithStdoutChannel(stdout chan<- string) ContainerOption {
	return func(c *Container) {
		c.stdout = stdout
	}
}

func WithStderrChannel(stderr chan<- string) ContainerOption {
	return func(c *Container) {
		c.stderr = stderr
	}
}

func WithUser(user string) ContainerOption {
	return func(c *Container) {
		c.user = user
	}
}
