package ingestncc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHandlerPanicsOnNilParser(t *testing.T) {
	assert.Panics(t, func() { NewHandler(nil) })
}
