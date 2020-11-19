package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNeverFail(t *testing.T) {
	assert.False(t, false, "Should never fail!")
}
