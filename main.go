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
	d := NewDiscoverer()

	modules, err := d.GetModules()
	if err != nil {
		return fmt.Errorf("getting modules: %w", err)
	}

	fmt.Printf("%+v", modules)

	return nil
}