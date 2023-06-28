package file

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

var Fs afero.Fs

func SetFs(fs afero.Fs) {
	Fs = fs
}

func ExistsAndIsFile(path string) bool {
	fileInfo, err := Fs.Stat(path)
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
