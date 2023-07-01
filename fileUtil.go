package file

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
)

type Logger interface {
	Fatal(args ...interface{})
	Panic(args ...interface{})
	Error(args ...interface{})
	Warn(args ...interface{})
	Info(args ...interface{})
	Debug(args ...interface{})
	Trace(args ...interface{})
	Print(args ...interface{})
}

var log Logger

func SetLogger(l Logger) {
	log = l
}

var fs afero.Fs

func SetFs(newFs afero.Fs) {
	fs = newFs
}

func ExistsAndIsFile(path string) bool {
	fileInfo, err := fs.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		fmt.Print(err.Error())
		return false
	}

	if fileInfo.IsDir() {
		return false
	}
	return true
}

func CreateMockFileSystem(t *testing.T, Fs afero.Fs) (string, string) {
	usr, err := user.Current()
	assert.Nil(t, err)

	// Creating a current directory in the in memory file system that represents the current directory.
	err = Fs.MkdirAll(usr.HomeDir, 777)
	assert.Nil(t, err)

	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	assert.Nil(t, err)
	// Creating the users home directory in the in memory file system the same as the current users home directory.
	err = Fs.MkdirAll(currentDir, 777)
	return currentDir, usr.HomeDir
}

func CopyDirectoryContents(source, destination string) error {
	err := afero.Walk(fs, source, func(path string, info os.FileInfo, err error) error {
		var relPath = strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return fs.Mkdir(filepath.Join(destination, relPath), 0755)
		} else {
			var data, err1 = afero.ReadFile(fs, filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return afero.WriteFile(fs, filepath.Join(destination, relPath), data, 0777)
		}
	})
	return err
}

func FileExists(path string) bool {
	_, err := fs.Stat(path)
	return !os.IsNotExist(err)
}
