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

	modules, err := d.ParseModules(listOutput)
	if err != nil {
		fmt.Printf("Failed to parse modules: %s\n", err)
	}
	for _, module := range modules {
		fmt.Printf("%+v\n", module)
	}

	return nil
}
