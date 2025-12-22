package main

import (
	"log"
	"os/exec"
)

func main() {
	log.Println("▶️ Starting application...")

	cmd := exec.Command("go", "run", "main.go")
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
