package main

import (
	"fmt"
)

func main() {
	name := "./upgorade-recipe.json"
	recipe, err := loadRecipe(name)
	if err != nil {
		fmt.Println("Failed to load recipe:", err)
		return
	}
	recipe.run()
}
