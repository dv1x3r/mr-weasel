package utils

import (
	"io"
	"os"
)

func CopyCrossDevice(src string, dst string) error {
	fin, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fin.Close()

	fout, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fout.Close()

	_, err = io.Copy(fout, fin)
	return err
}

func MoveCrossDevice(src string, dst string) error {
	err := CopyCrossDevice(src, dst)
	if err != nil {
		return err
	}
	return os.Remove(src)
}
