package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	assert.Equal(t, "Skipped", "Skipped", "There's no need to test the func main in this library. It's not a part of any execution program. Main in this repository is only a simple test utilities.")
}
