package main

import (
	"log"
	"os"

	cmd "github.com/chimera-kube/chimera-admission/cmd/chimera"
)

func main() {
	app := cmd.NewApp()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
