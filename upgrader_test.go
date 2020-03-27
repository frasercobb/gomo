package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UpgradeReturnsErrorFromExecutor(t *testing.T) {
	wantError := fmt.Errorf("an-error-from-executor")
	mockExecutor := MockExecutor{RunError: wantError}
	u := NewUpgrader(
		WithUpgradeExecutor(&mockExecutor),
	)

	err := u.UpgradeModules([]Module{{Name: "foo/bar"}})

	assert.Contains(t, err.Error(), fmt.Sprintf(`upgrading module "foo/bar": %s`, wantError.Error()))
}

func Test_UpgradeCallsExecutorRunWithTheCorrectArguments(t *testing.T) {
	mockExecutor := MockExecutor{}
	u := NewUpgrader(
		WithUpgradeExecutor(&mockExecutor),
	)

	aModuleName := "frasercobb/gomo"
	modules := []Module{
		{Name: aModuleName},
	}
	err := u.UpgradeModules(modules)
	require.NoError(t, err)

	runCalls := mockExecutor.RunCalls
	require.Len(t, runCalls, 1)

	args := fmt.Sprintf("get %s", aModuleName)
	assert.Equal(t, runCalls[0], RunCall{
		Command: "go",
		Args:    args,
	})
}
