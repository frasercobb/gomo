package main

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
)

func Test_CreateSelectOptionsColoursPatch(t *testing.T) {
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

func Test_CreateSelectOptionsColoursMinor(t *testing.T) {
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

func Test_CreateSelectOptionsGroupsByUpgradeType(t *testing.T) {
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

func Test_AskForUpgradesReturnsErrorWhenNoModulesGiven(t *testing.T) {
	p := NewPrompter()

	_, err := p.AskForUpgrades([]Module{})

	assert.Contains(t, err.Error(), "unable to create upgrade prompt: ")
}
