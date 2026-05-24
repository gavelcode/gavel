package ui_test

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/ui"
)

func TestFormatDurationSecondsOnly(t *testing.T) {
	assert.Equal(t, "0s", ui.FormatDuration(0))
	assert.Equal(t, "5s", ui.FormatDuration(5*time.Second))
	assert.Equal(t, "59s", ui.FormatDuration(59*time.Second))
}

func TestFormatDurationWithMinutes(t *testing.T) {
	assert.Equal(t, "1m 00s", ui.FormatDuration(60*time.Second))
	assert.Equal(t, "1m 30s", ui.FormatDuration(90*time.Second))
	assert.Equal(t, "5m 05s", ui.FormatDuration(5*time.Minute+5*time.Second))
}

func TestNewSpinner(t *testing.T) {
	var buf bytes.Buffer
	spinner := ui.NewSpinner(&buf)
	require.NotNil(t, spinner)
}

func TestNewSpinnerWithMessage(t *testing.T) {
	var buf bytes.Buffer
	spinner := ui.NewSpinnerWithMessage(&buf, "Loading...")
	require.NotNil(t, spinner)
}

func TestSpinnerRunAndStop(t *testing.T) {
	var buf bytes.Buffer
	spinner := ui.NewSpinnerWithMessage(&buf, "working")

	go spinner.Run()
	time.Sleep(250 * time.Millisecond)
	spinner.Stop()

	assert.True(t, buf.Len() > 0)
}

func TestSpinnerElapsed(t *testing.T) {
	var buf bytes.Buffer
	spinner := ui.NewSpinner(&buf)

	go spinner.Run()
	time.Sleep(150 * time.Millisecond)
	spinner.Stop()

	assert.GreaterOrEqual(t, spinner.Elapsed(), time.Duration(0))
}

func TestSpinnerRunWriteError(t *testing.T) {
	w := &failAfterNWriter{max: 1}
	spinner := ui.NewSpinner(w)

	go spinner.Run()
	time.Sleep(350 * time.Millisecond)
	spinner.Stop()
}

type failAfterNWriter struct {
	n   int
	max int
}

func (f *failAfterNWriter) Write(p []byte) (int, error) {
	f.n++
	if f.n > f.max {
		return 0, errors.New("write failed")
	}
	return len(p), nil
}

func TestIsTerminal_NonTerminalWriter(t *testing.T) {
	var buf bytes.Buffer
	assert.False(t, ui.IsTerminal(&buf))
}

func TestIsTerminal_FdWriterNonTTY(t *testing.T) {
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, reader.Close())
		require.NoError(t, writer.Close())
	})

	assert.False(t, ui.IsTerminal(writer))
}

func TestThemeGavel(t *testing.T) {
	theme := ui.ThemeGavel()
	require.NotNil(t, theme)
}
