package main

func main() {
	recipe := &Recipe{
		Source{
			"http://files.kaoriya.net/vim/snapshots/latest.json",
			Selectors {
				"vim74w32",
				"vim74w64",
			},
		},
		[]string {
			"gvim.exe",
			"vim.exe",
		},
	}
	recipe.run()
}
