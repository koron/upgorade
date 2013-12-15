package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

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

func LoadRecipe(name string) (*Recipe, error) {
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
	archives, err := fetchArchives(recipe.Source.Url)
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

func (recipe *Recipe) Run() {
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
