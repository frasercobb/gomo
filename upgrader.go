package main

import "fmt"

type Upgrader struct {
	Executor Executor
}

type UpgraderOption func(*Upgrader)

func NewUpgrader(options ...UpgraderOption) *Upgrader {
	u := &Upgrader{
		Executor: nil,
	}

	for _, option := range options {
		option(u)
	}

	return u
}

func WithUpgradeExecutor(executor Executor) UpgraderOption {
	return func(u *Upgrader) {
		u.Executor = executor
	}
}

func (u *Upgrader) UpgradeModules(modules []Module) error {
	for _, mod := range modules {
		if err := u.upgradeModule(mod); err != nil {
			return fmt.Errorf("upgrading module %q: %w", mod.Name, err)
		}
	}
	return nil
}

func (u *Upgrader) upgradeModule(module Module) error {
	_, err := u.Executor.Run("go", "get", module.Name)
	if err != nil {
		return err
	}

	return nil
}
