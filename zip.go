package file

import (
	"archive/zip"
	"fmt"
	"github.com/flightx31/exception"
	"github.com/spf13/afero"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipSource copied from here: https://gosamples.dev/zip-file/
func ZipSource(source, target string, skipContainingFolder bool) error {
	// 1. Create a ZIP file and zip.Writer
	f, err := fs.Create(target)
	if err != nil {
		return err
	}
	info, err := fs.Stat(source)
	if err != nil {
		return err
	}
	skipSourceDirectory := skipContainingFolder && info.IsDir()

	writer := zip.NewWriter(f)
	defer writer.Close()

	// 2. Go through all the files of the source
	return afero.Walk(fs, source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == source && skipSourceDirectory {
			return nil
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		p, err := filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}

		// 4. Set relative path of a file as the header name
		if skipSourceDirectory {
			base := filepath.Base(source)
			header.Name = strings.Replace(p, base, "", 1)
		} else {
			header.Name, err = filepath.Rel(filepath.Dir(source), path)
			if err != nil {
				return err
			}
		}

		if info.IsDir() {
			header.Name += "/"
		}

		// 5. Create writer for the file header and save content of the file
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := fs.Open(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(headerWriter, f)
		return err
	})
}

// Unzip - unzips a zip file. Resource: https://golang.cafe/blog/golang-unzip-file-example.html
func Unzip(sourceZip string, destDirectory string, programName string) error {
	// Create temporary directory in the default location for temporary files
	tempDir, err := afero.TempDir(fs, "", programName)
	defer func() {
		exception.LogError(fs.RemoveAll(tempDir))
	}()

	if err != nil {
		return err
	}
	log.Info(`(UPGRADE): creating temporary directory at "`, tempDir, `"`)

	archive, err := zip.OpenReader(sourceZip)
	if err != nil {
		return err
	}
	defer func() {
		exception.LogError(archive.Close())
	}()

	// Unzip all the files/folders in the zip archive
	for _, zFile := range archive.File {

		// skip this odd (useless?) folder
		if strings.HasPrefix(zFile.Name, "__MACOSX") {
			// Resource: https://wpguru.co.uk/2013/10/how-to-remove-__macosx-from-zip-archives/
			continue
		}

		log.Trace(`(UPGRADE): unzipping file to teporary directory"`, zFile.Name, `" to "`, tempDir, `"`)
		err = unzipSingleFile(zFile, tempDir)

		if err != nil {
			// if there is an error unzipping a file return the error which will abort the unzipping and delete the temporary file
			return err
		}
	}

	// create destination directory if not exists
	err = fs.MkdirAll(destDirectory, os.ModePerm)

	if err != nil {
		return err
	}

	// copy all the unzipped files to the new directory
	err = CopyDirectoryContents(tempDir, destDirectory)

	if err != nil {
		return err
	}

	return nil
}

func unzipSingleFile(zFile *zip.File, destination string) error {
	// todo: add code to check for duplicates when unzipping and throw error if found.

	filePath := filepath.Join(destination, zFile.Name)

	// Check for ZipSlip: https://snyk.io/research/zip-slip-vulnerability
	if !strings.HasPrefix(filePath, destination) {
		return fmt.Errorf("%s: illegal zFile path", filePath)
	}

	if zFile.FileInfo().IsDir() {
		err := fs.MkdirAll(filePath, os.ModePerm)

		if err != nil {
			return err
		}

		return nil
	}

	_, err := fs.Create(filePath)

	if err != nil {
		return err
	}

	// unzip zFile
	destFile, err := fs.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}

	fileInArchive, err := zFile.Open()
	if err != nil {
		if err := fileInArchive.Close(); err != nil {
			panic(err)
		}
		return err
	}

	_, err = io.Copy(destFile, fileInArchive)
	if err != nil {
		return err
	}

	// close open things.
	if err := fileInArchive.Close(); err != nil {
		panic(err)
	}

	return nil
}
