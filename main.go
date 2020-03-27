package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Encountered an error %s", err)
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

	p := NewPrompter()
	modulesToUpgrade, err := p.AskForUpgrades(modules)
	if err != nil {
		return fmt.Errorf("asking for which modules to upgrade: %w", err)
	}

	u := NewUpgrader(
		WithUpgradeExecutor(cmdExecutor),
	)
	if err := u.UpgradeModules(modulesToUpgrade); err != nil {
		return err
	}

	return nil
}
