package gozip

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

/*
CustomZip facility the process to zip and unzip files in golang

Usage:
	1 - Create a new zip file
		oZip := CustomZip()

		oZip.Create("./testes/arq.zip")

		oZip.Add("./testes/clientes.txt")
		oZip.Add("./testes/imagem.png")

		oZip.Close()

	2 - Unzip file
		obj := CustomZip()

		obj.Unzip("./novozip.zip", "./unzip")
*/
func CustomZip() customZip {
	return customZip{PermMode: 0666}
}

type customZip struct {
	ZipFile  *os.File
	writer   *zip.Writer
	PermMode fs.FileMode
}

// Create creates a new zip file
func (o *customZip) Create(file string) error {
	var (
		err error
	)

	err = os.MkdirAll(filepath.Dir(file), o.PermMode)
	if err != nil {
		return err
	}

	o.ZipFile, err = os.Create(file)
	if err != nil {
		return err
	}

	o.writer = zip.NewWriter(o.ZipFile)

	return nil
}

func removePrefix(s string) string {
	s = strings.ReplaceAll(s, "\\", "/")

	if s[0:0+2] == "//" || s[0:0+2] == "./" {
		s = s[2:]
	} else if s[1:1+2] == ":/" {
		s = s[3:]
	}

	return s
}

//Add add a single file or a tree directory into the zip file
func (o *customZip) Add(originFile string) error {

	f, err := os.Stat(originFile)
	if err != nil {
		return err
	} else if f.IsDir() {
		err = filepath.WalkDir(originFile, o.walkFunc)

		return err
	}

	err = o.addFile(originFile)

	return err
}

func (o *customZip) walkFunc(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return nil
	}

	path = strings.ReplaceAll(path, "\\", "/")
	p := removePrefix(path)

	if d.IsDir() {
		_, err = o.writer.Create(p + "/")
		return err
	}

	err = o.addFile(path)
	return err
}

func (o *customZip) addFile(originFile string) error {

	oF, err := os.Open(originFile)
	if err != nil {
		return err
	}
	defer oF.Close()

	zF, err := o.writer.Create(removePrefix(originFile))
	if err != nil {
		return err
	}

	_, err = io.Copy(zF, oF)
	return err
}

//Close closes a zip file
func (o *customZip) Close() error {
	o.writer.Close()
	return o.ZipFile.Close()
}

//Unzip unzip zipFile to destFolder
func (o *customZip) Unzip(zipFile string, destFolder string) error {

	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}

	defer r.Close()

	err = os.MkdirAll(destFolder, o.PermMode)
	if err != nil {
		return err
	}

	for _, f := range r.File {

		destFile := destFolder + "/" + f.Name

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(destFile, o.PermMode)
			if err != nil {
				return err
			}

			continue
		} else {
			err = os.MkdirAll(filepath.Dir(destFile), o.PermMode)
			if err != nil {
				return err
			}
		}

		outFile, err := os.Create(destFile)
		if err != nil {
			return err
		}

		zipReader, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, zipReader)
		if err != nil {
			return err
		}

		outFile.Close()
		zipReader.Close()
	}

	return nil
}
