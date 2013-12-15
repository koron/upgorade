package main

import (
	"./common"
)

func main() {
	recipe := &common.Recipe{
		common.Source{
			"http://files.kaoriya.net/vim/snapshots/latest.json",
			common.Selectors {
				"vim74w32",
				"vim74w64",
			},
		},
		[]string {
			"gvim.exe",
			"vim.exe",
		},
	}
	recipe.Run()
}
