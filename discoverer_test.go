package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/frasercobb/gomo/mock"
)

func Test_ListModulesCallsExecutorRun(t *testing.T) {
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(&mockExecutor)

	_, err := d.ListModules()
	require.NoError(t, err)

	runCalls := mockExecutor.RunCalls
	require.Len(t, runCalls, 1)

	listArgs := "list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' -m all"
	assert.Equal(t, runCalls[0], mock.RunCall{
		Command: "go",
		Args:    listArgs,
	})
}

func Test_ListModulesReturnsErrorFromExecutor(t *testing.T) {
	wantError := fmt.Errorf("an-error-from-executor")
	mockExecutor := mock.Executor{RunError: wantError}
	d := NewDiscoverer(&mockExecutor)

	_, err := d.ListModules()

	assert.Error(t, wantError, err)
}

func Test_ListModulesReturnsModules(t *testing.T) {
	moduleName := "a-module-name"
	wantModules := []Module{
		{
			Name: moduleName,
		},
	}

	modulesListOutput := modulesToListFormat(wantModules...)
	mockExecutor := mock.Executor{
		CommandOutput: modulesListOutput,
	}

	d := Discoverer{
		Executor: &mockExecutor,
	}

	moduleOutput, err := d.ListModules()
	require.NoError(t, err)

	assert.Equal(t, modulesListOutput, moduleOutput)
}

func Test_ListModulesHandlesLatestModules(t *testing.T) {
	commandOutput := []string{
		"go: finding golang.org/x/sync latest",
		"go: finding golang.org/x/net latest",
		"go: finding gopkg.in/tomb.v1 latest",
	}

	result := strings.Join(commandOutput, "\n")
	mockExecutor := mock.Executor{
		CommandOutput: result,
	}

	d := Discoverer{
		Executor: &mockExecutor,
	}

	moduleOutput, err := d.ListModules()
	require.NoError(t, err)

	assert.Equal(t, result, moduleOutput)
}

func Test_ParseModulesReturnsErrorWhenInvalidModuleRegex(t *testing.T) {
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(&mockExecutor)
	d.ModuleRegex = "not a valid regex ("

	_, err := d.ParseModules("")
	require.Error(t, err)

	assert.Contains(t, err.Error(), "error parsing regexp")
}

func Test_ParseModulesReturnsErrorWhenNotAllMatched(t *testing.T) {
	output := "example.com/a/module: 1.0.0 ->"
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(&mockExecutor)

	_, err := d.ParseModules(output)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "regex was not able to find all matches")
}

func Test_ParseModulesReturnsErrorWhenFromIsNotAValidSemver(t *testing.T) {
	wantModule := Module{
		Name:      "a-module-name",
		ToVersion: semver.Version{Major: 0, Minor: 4, Patch: 5},
	}
	output := fmt.Sprintf("%s: %s -> %s", wantModule.Name, "not-a-from-version", wantModule.ToVersion)
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(&mockExecutor)

	_, err := d.ParseModules(output)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "parsing from version:")
}

func Test_ParseModulesReturnsErrorWhenToIsNotAValidSemver(t *testing.T) {
	wantModule := Module{
		Name:        "a-module-name",
		FromVersion: semver.Version{Major: 0, Minor: 4, Patch: 5},
	}
	output := fmt.Sprintf("%s: %s -> %s", wantModule.Name, wantModule.FromVersion, "not-a-valid-semver")
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(&mockExecutor)

	_, err := d.ParseModules(output)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "parsing to version:")
}

func Test_ParseModulesReturnsExpectedModule(t *testing.T) {
	wantModule := Module{
		Name:        "a-module-name",
		FromVersion: semver.Version{Major: 0, Minor: 3, Patch: 5},
		ToVersion:   semver.Version{Major: 0, Minor: 4, Patch: 5},
	}
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(&mockExecutor)

	moduleListOutput := modulesToListFormat(wantModule)
	modules, err := d.ParseModules(moduleListOutput)
	require.NoError(t, err)
	require.Len(t, modules, 1)

	assert.Equal(t, wantModule, modules[0])
}

func Test_ParseModulesReturnsExpectedModules(t *testing.T) {
	wantModules := []Module{
		{
			Name:        "a-module-name",
			FromVersion: semver.Version{Major: 0, Minor: 3, Patch: 5},
			ToVersion:   semver.Version{Major: 0, Minor: 4, Patch: 5},
		},
		{
			Name:        "another-module-name",
			FromVersion: semver.Version{Major: 1, Minor: 2, Patch: 3},
			ToVersion:   semver.Version{Major: 1, Minor: 2, Patch: 2},
		},
	}
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(&mockExecutor)
	moduleListOutput := modulesToListFormat(wantModules...)
	modules, err := d.ParseModules(moduleListOutput)
	require.NoError(t, err)
	require.Len(t, modules, 2)

	assert.Equal(t, wantModules, modules)
}

func modulesToListFormat(modules ...Module) string {
	var result []string
	for _, module := range modules {
		result = append(result, fmt.Sprintf("%s: %s -> %s", module.Name, module.FromVersion, module.ToVersion))
	}
	return strings.Join(result, "\n")
}