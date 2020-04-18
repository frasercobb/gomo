package main

import (
	"bytes"
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
	buf := new(bytes.Buffer)
	c, _, err := vt10x.NewVT10XConsole(expect.WithStdout(buf), expect.WithDefaultTimeout(1*time.Second))
	require.Nil(t, err)
	defer c.Close()

	stdio := terminal.Stdio{Out: c.Tty(), In: c.Tty(), Err: c.Tty()}

	donec := make(chan struct{})
	go func() {
		defer close(donec)
		_, err := c.ExpectString("Which modules do you want to upgrade?")
		assert.NoError(t, err)
		_, err = c.Send(string(terminal.KeyArrowDown))
		assert.NoError(t, err)
		_, err = c.SendLine(" ")
		assert.NoError(t, err)
		_, err = c.ExpectEOF()
		assert.NoError(t, err)
	}()

	p := NewPrompter(
		WithStdio(stdio),
	)

	modules := []Module{
		{
			Name:         "minor/upgrade",
			FromVersion:  semver.MustParse("0.1.1"),
			ToVersion:    semver.MustParse("0.2.1"),
			MinorUpgrade: true,
		},
	}
	gotModules, err := p.AskForUpgrades(modules)
	require.NoError(t, err)

	c.Tty().Close()
	<-donec

	assert.Len(t, gotModules, 1)
	assert.Contains(t, gotModules, modules[0])
}
