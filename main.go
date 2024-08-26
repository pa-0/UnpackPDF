package main

import (
	"fmt"

	"github.com/mattheusrocha2/mergeJPG-PDF/app"
)

func main() {
	app.FindFolder("tmp\\")
	fmt.Println(" ")
	fmt.Println("Processamento concluído. Pressione Enter para sair...")
	var input string
	fmt.Scanln(&input)

}
