package program

import (
	"errors"
	"io/fs"
	"os"
)

func checkFileExists(file string) error {
	if _, err := os.Stat(file); errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}
