package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EndToEndHappyPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	err := run()

	assert.NoError(t, err)
}
