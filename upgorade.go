package main

import (
	"./common"
	"fmt"
)

func main() {
	name := "./upgorade-recipe.json"
	recipe, err := common.LoadRecipe(name)
	if err != nil {
		fmt.Println("Failed to load recipe:", err)
		return
	}
	recipe.Run()
}
