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

	for _, mod := range modules {
		changelog, err := d.GetChangelog(mod)
		if err != nil {
			fmt.Printf("Error: %s", err)
			continue
		}
		fmt.Printf("Changelog for %+v: %s", mod, changelog)
	}

	return nil
}