package main

import (
	"fmt"

	"github.com/pa-0/UnpackPDF/app"
)

func main() {
	app.UnpackPDF("tmp\\")
	fmt.Println(" ")
	fmt.Println("Extraction Complete. Press enter to exit...")
	var input string
	fmt.Scanln(&input)

}
