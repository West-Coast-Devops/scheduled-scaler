package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNeverFail(t *testing.T) {
	assert.False(t, false, "Should never fail!")
}
