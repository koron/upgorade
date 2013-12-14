package main

import (
	"archive/zip"
	"crypto/sha1"
	"debug/pe"
	"encoding/hex"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

func extractOne(name string, zf *zip.File) error {
	r, err := zf.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(name)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	return err
}

func rotateName(name string, index int) (r string) {
	r = name
	if index == 0 {
		return
	}
	ext := path.Ext(name)
	basename := name[:len(name)-len(ext)]
	return basename + "." + strconv.Itoa(index) + ext
}

func rotate(name string, max int) (err error) {
	last := rotateName(name, max)
	err = os.Remove(last)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	for i := max - 1; i >= 0; i -= 1 {
		curr := rotateName(name, i)
		err = os.Rename(curr, last)
		if err != nil && !os.IsNotExist(err) {
			return
		}
		last = curr
	}
	return nil
}

func stripPath(name string) string {
	s := strings.Split(name, "/")
	return path.Join(s[1:]...)
}

func calcSha1(name string) (*string, error) {
	r, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	h := sha1.New()
	_, err = io.Copy(h, r)
	if err != nil {
		return nil, err
	}
	hs := hex.EncodeToString(h.Sum(nil))

	return &hs, nil
}

func loadArch(name string) (arch string, err error) {
	f, err := pe.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}
	defer f.Close()

	switch f.FileHeader.Machine {
	case 0x014c:
		arch = "x86"
	case 0x8664:
		arch = "amd64"
	}
	return
}
