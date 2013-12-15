package common

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	URL "net/url"
	"os"
	"strings"
)

type File struct {
	Name string `json:"name"`
	Size uint64 `json:"size"`
	Sha1 string `json:"sha1"`
}

type Archive struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Contents []File `json:"contents"`
}

func (a *Archive) filename() (*string, error) {
	url, err := URL.Parse(a.Url)
	if err != nil {
		return nil, err
	}
	s := strings.Split(url.Path, "/")
	var name string
	if len(s) > 0 {
		name = s[len(s)-1]
	} else {
		name = ""
	}
	if len(name) == 0 {
		name = a.Name + ".zip"
	}
	return &name, nil
}

func (a *Archive) get() (*string, error) {
	resp, err := http.Get(a.Url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	name, err := a.filename()
	if err != nil {
		return nil, err
	}
	f, err := os.Create(*name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		f.Close()
		os.Remove(*name)
		return nil, err
	}
	return name, nil
}

func (a *Archive) update() error {
	name, err := a.get()
	if err != nil {
		return err
	}
	defer os.Remove(*name)

	zr, err := zip.OpenReader(*name)
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, zf := range zr.File {
		if zf.Mode().IsDir() {
			continue
		}

		outname := stripPath(zf.FileHeader.Name)
		err = rotate(stripPath(zf.FileHeader.Name), 5)
		if err != nil {
			return err
		}

		err = extractOne(outname, zf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Archive) hasUpdate() (bool, error) {
	for _, f := range a.Contents {
		name := stripPath(f.Name)
		sha1, err := calcSha1(name)
		if err != nil {
			if os.IsNotExist(err) {
				return true, nil
			}
			return false, err
		}
		if sha1 == nil || *sha1 != f.Sha1 {
			return true, nil
		}
	}
	return false, nil
}

func fetchArchives(url string) ([]Archive, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status %d for %s",
			resp.StatusCode, url)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var archives []Archive
	err = json.Unmarshal(b, &archives)
	if err != nil {
		return nil, err
	}
	return archives, nil
}
