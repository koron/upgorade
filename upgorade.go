package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	URL "net/url"
	"os"
	"path"
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

type Selectors struct {
	X86   string `json:"x86"`
	Amd64 string `json:"amd64"`
}

type Source struct {
	Url       string `json:"url"`
	Selectors `json:"selectors"`
}

func (s Source) fetchArchives() ([]Archive, error) {
	resp, err := http.Get(s.Url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode,
			s.Url)
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

func (s Source) find(arch string) (name string, err error) {
	if arch == "x86" {
		name = s.Selectors.X86
	} else if arch == "amd64" {
		name = s.Selectors.Amd64
	} else {
		err = fmt.Errorf("uknown arch %s", arch)
	}
	return
}

type Recipe struct {
	Source  `json:"source"`
	Targets []string `json:"targets"`
}

func loadRecipe(name string) (*Recipe, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var recipe Recipe
	err = json.Unmarshal(b, &recipe)
	if err != nil {
		return nil, err
	}
	return &recipe, err
}

func (recipe *Recipe) arch() (name string, err error) {
	for _, target := range recipe.Targets {
		ext := path.Ext(target)
		if ext == ".exe" {
			name, err = loadArch(target)
			if err != nil || name != "" {
				return
			}
		}
	}
	// FIXME: check architecture of OS, or my self.
	return "x86", nil
}

func (recipe *Recipe) guessArchive(archives []Archive) (*Archive, error) {
	arch, err := recipe.arch()
	if err != nil {
		return nil, err
	}

	name, err := recipe.Source.find(arch)
	if err != nil {
		return nil, err
	}

	for _, a := range archives {
		if a.Name == name {
			return &a, nil
		}
	}
	return nil, fmt.Errorf("Unknown archive name %s", name)
}

func (recipe *Recipe) upgrade() (result bool, err error) {
	archives, err := recipe.Source.fetchArchives()
	if err != nil {
		return
	}
	archive, err := recipe.guessArchive(archives)
	if err != nil {
		return
	}
	hasUpdated, err := archive.hasUpdate()
	if err != nil {
		return
	}
	if !hasUpdated {
		return
	}
	err = archive.update()
	if err == nil {
		result = true
	}
	return
}

func (recipe *Recipe) run() {
	upgraded, err := recipe.upgrade()
	if err != nil {
		fmt.Println("Upgrade failed:", err)
		return
	}
	if upgraded {
		fmt.Println("upgraded successfully")
	} else {
		fmt.Println("no upgrade")
	}
}
