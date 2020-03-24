package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Flags()
	if err := run(); err != nil {
		log.Printf("Encountered an error %s", err)
	}
}

func run() error {
	cmdExecutor := NewCommandExecutor()
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	d := NewDiscoverer(
		WithExecutor(cmdExecutor),
		WithHTTPClient(&client),
	)

	modules, err := d.GetModules()
	if err != nil {
		return fmt.Errorf("getting modules: %w", err)
	}

	fmt.Printf("Name\t\t\t\tCurrent\tUpgrade\tChangelog\n")
	for _, mod := range modules {
		changelog, _ := d.GetChangelog(mod)
		fmt.Printf("%s\t%s\t%s\t%s\n", mod.Name, mod.FromVersion, mod.ToVersion, changelog)
	}
	return nil
}
