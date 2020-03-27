package main

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"gotest.tools/assert"
)

func Test_CreateSelectOptions(t *testing.T) {
	modules := []Module{
		{
			Name:        "foo/bar",
			FromVersion: semver.MustParse("0.0.1"),
			ToVersion:   semver.MustParse("0.0.2"),
		},
		{
			Name:        "frasercobb/gomo",
			FromVersion: semver.MustParse("1.2.3"),
			ToVersion:   semver.MustParse("1.3.4"),
		},
	}

	result := createSelectOptions(modules)

	assert.DeepEqual(t, result, []string{
		"foo/bar 0.0.1 -> 0.0.2",
		"frasercobb/gomo 1.2.3 -> 1.3.4",
	})
}

func Test_AskForUpgradesReturnsErrorWhenNoModulesGiven(t *testing.T) {
	p := NewPrompter()

	_, err := p.AskForUpgrades([]Module{})

	assert.ErrorContains(t, err, "unable to create upgrade prompt: ")
}
