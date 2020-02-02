package main

import (
	"fmt"
	"log"
)

func main() {
	log.Flags()
	err := run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	cmdExecutor := CommandExecutor{}

	d := NewDiscoverer(&cmdExecutor)

	listOutput, err := d.ListModules()
	if err != nil {
		return fmt.Errorf("listing modules: %w", err)
	}

	fmt.Print(listOutput)

	return nil
}
