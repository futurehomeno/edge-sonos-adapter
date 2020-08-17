package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func CopyFile(src, dst string) error {

	closeFile := func(c io.Closer) {
		if c == nil {
			return
		}
		if err := c.Close(); err != nil {
			log.Error(err)
		}
	}

	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer closeFile(source)

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer closeFile(destination)
	_, err = io.Copy(destination, source)
	return err
}
