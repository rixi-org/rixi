package main

import (
	"fmt"
)

func generate(name string) {
	fmt.Printf("generating model %s...\n", name)
	generateModel(name)

	fmt.Printf("generating view %s...\n", name)
	generateView(name)

	fmt.Printf("generating controller %s...\n", name)
	generateController(name)

	fmt.Println("done")
}
