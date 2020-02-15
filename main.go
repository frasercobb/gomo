package main

import (
	"fmt"
	"log"
)

func main() {
	log.Flags()
	err := run()
	if err != nil {
		log.Printf("Encountered an error %s", err)
	}
}

func run() error {
	cmdExecutor := CommandExecutor{}

	d := NewDiscoverer(&cmdExecutor)

	modules, err := d.GetModules()
	if err != nil {
		return fmt.Errorf("getting modules: %w", err)
	}

	modules, err := d.ParseModules(listOutput)
	if err != nil {
		fmt.Printf("Failed to parse modules: %s\n", err)
	}

	return nil
}

		if module.MinorUpgrade {
			fmt.Printf("Minor: %+v\n", module)
			continue
		}
	}

	return nil
}
