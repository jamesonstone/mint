package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type initFile struct {
	path    string
	content []byte
}

var initDirectories = []string{".github", filepath.Join(".github", "workflows")}

func checkInitPaths(files []initFile) error {
	for _, directory := range initDirectories {
		if err := checkInitDirectory(directory); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	for _, file := range files {
		if _, err := os.Lstat(file.path); !os.IsNotExist(err) {
			if err != nil {
				return err
			}
			return fmt.Errorf("cannot initialize: %s already exists", file.path)
		}
	}
	return nil
}

func checkInitDirectory(directory string) error {
	info, err := os.Lstat(directory)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("cannot initialize: %s must be a directory", directory)
	}
	return nil
}

func installInitFiles(files []initFile) (result error) {
	var created []string
	defer func() {
		if result == nil {
			return
		}
		for index := len(created) - 1; index >= 0; index-- {
			if err := os.Remove(created[index]); err != nil && !os.IsNotExist(err) {
				result = errors.Join(result, err)
			}
		}
	}()
	for _, directory := range initDirectories {
		err := os.Mkdir(directory, 0755)
		if err == nil {
			created = append(created, directory)
			continue
		}
		if !os.IsExist(err) {
			return err
		}
		if err := checkInitDirectory(directory); err != nil {
			return err
		}
	}
	for _, file := range files {
		target, err := os.OpenFile(file.path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		created = append(created, file.path)
		_, err = target.Write(file.content)
		if err = errors.Join(err, target.Close()); err != nil {
			return fmt.Errorf("write %s: %w", file.path, err)
		}
	}
	return nil
}
