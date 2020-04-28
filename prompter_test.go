package main

import (
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/Masterminds/semver/v3"
	"github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateSelectOptions_ColoursPatch(t *testing.T) {
	modules := []Module{
		{
			Name:         "frasercobb/gomo",
			FromVersion:  semver.MustParse("1.2.3"),
			ToVersion:    semver.MustParse("1.2.4"),
			PatchUpgrade: true,
		},
	}
	result := createSelectOptions(modules)

	assert.Equal(t, []string{
		"\x1b[32mfrasercobb/gomo 1.2.3 -> 1.2.4\x1b[0m",
	}, result)
}

func Test_CreateSelectOptions_ColoursMinor(t *testing.T) {
	modules := []Module{
		{
			Name:         "foo/bar",
			FromVersion:  semver.MustParse("0.1.0"),
			ToVersion:    semver.MustParse("0.2.0"),
			MinorUpgrade: true,
		},
	}
	result := createSelectOptions(modules)

	assert.Equal(t, []string{
		"\x1b[34mfoo/bar 0.1.0 -> 0.2.0\x1b[0m",
	}, result)
}

func Test_CreateSelectOptions_GroupsByUpgradeType(t *testing.T) {
	modules := []Module{
		{
			Name:         "minor/upgrade",
			FromVersion:  semver.MustParse("0.1.1"),
			ToVersion:    semver.MustParse("0.2.1"),
			MinorUpgrade: true,
		},
		{
			Name:         "patch/upgrade",
			FromVersion:  semver.MustParse("0.0.1"),
			ToVersion:    semver.MustParse("0.0.2"),
			PatchUpgrade: true,
		},
	}
	result := createSelectOptions(modules)

	assert.Equal(t, []string{
		"\x1b[32mpatch/upgrade 0.0.1 -> 0.0.2\x1b[0m",
		"\x1b[34mminor/upgrade 0.1.1 -> 0.2.1\x1b[0m",
	}, result)
}

func Test_AskForUpgrades_ReturnsErrorWhenNoModulesGiven(t *testing.T) {
	p := NewPrompter()

	_, err := p.AskForUpgrades([]Module{})

	assert.Contains(t, err.Error(), "unable to get module choices: ")
}

func Test_AskForUpgrades_CanSelectAModule(t *testing.T) {
	modules := []Module{
		{
			Name:         "minor/upgrade",
			FromVersion:  semver.MustParse("0.1.1"),
			ToVersion:    semver.MustParse("0.2.1"),
			MinorUpgrade: true,
		},
	}
	sendInputs := func(c *expect.Console) {
		_, _ = c.ExpectString("Which modules do you want to upgrade?")
		_, _ = c.Send(string(terminal.KeyArrowDown))
		_, _ = c.SendLine(" ")
		_, _ = c.ExpectEOF()
	}
	gotModules, err := RunPrompterCLITest(t, modules, sendInputs)
	require.NoError(t, err)

	assert.Len(t, gotModules, 1)
	assert.Contains(t, gotModules, modules[0])
}

func RunPrompterCLITest(t *testing.T, given []Module, sendInputs func(*expect.Console)) ([]Module, error) {
	c, _, err := vt10x.NewVT10XConsole(expect.WithDefaultTimeout(100 * time.Millisecond))
	require.Nil(t, err)
	defer c.Close()
	defer c.Tty().Close()

	stdio := terminal.Stdio{Out: c.Tty(), In: c.Tty(), Err: c.Tty()}
	p := NewPrompter(WithStdio(stdio))

	errCh := make(chan error)
	go sendInputs(c)

	modulesCh := make(chan []Module)
	go func() {
		gotModules, err := p.AskForUpgrades(given)
		if err != nil {
			errCh <- err
		}
		modulesCh <- gotModules
	}()

	select {
	case gotModules := <-modulesCh:
		return gotModules, nil
	case err := <-errCh:
		return nil, err
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout during prompter CLI test")
		return nil, nil
	}
}
