package watch

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRegisterFlagsBindsAllFlags(t *testing.T) {
	cmd := &cobra.Command{}
	opts := &Options{}

	RegisterFlags(cmd, opts)

	assert.NotNil(t, cmd.Flags().Lookup("debounce"))
	assert.NotNil(t, cmd.Flags().Lookup("languages"))
	assert.NotNil(t, cmd.Flags().Lookup("workspace"))
}
