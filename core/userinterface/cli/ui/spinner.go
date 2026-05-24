package ui

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
)

const secondsPerMinute = 60

const spinnerTickInterval = 100 * time.Millisecond

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	w       io.Writer
	message string
	stop    chan struct{}
	stopped chan struct{}
	once    sync.Once
	start   time.Time
}

func NewSpinner(w io.Writer) *Spinner {
	return &Spinner{
		w:       w,
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
		start:   time.Now(),
	}
}

func NewSpinnerWithMessage(w io.Writer, message string) *Spinner {
	return &Spinner{
		w:       w,
		message: message,
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
		start:   time.Now(),
	}
}

func (s *Spinner) Run() {
	ticker := time.NewTicker(spinnerTickInterval)
	defer ticker.Stop()
	defer close(s.stopped)

	frame := 0
	for {
		select {
		case <-s.stop:
			_, _ = fmt.Fprint(s.w, "\r\033[K")
			return
		case <-ticker.C:
			elapsed := time.Since(s.start).Truncate(time.Second)
			spin := GoldBar.Render(spinnerFrames[frame%len(spinnerFrames)])
			var err error
			if s.message != "" {
				_, err = fmt.Fprintf(s.w, "\r  %s %s %s", spin, Dim.Render(s.message), Dim.Render(FormatDuration(elapsed)))
			} else {
				_, err = fmt.Fprintf(s.w, "\r  %s %s", spin, Dim.Render(FormatDuration(elapsed)))
			}
			if err != nil {
				return
			}
			frame++
		}
	}
}

func (s *Spinner) Stop() {
	s.once.Do(func() { close(s.stop) })
	<-s.stopped
}

func (s *Spinner) Elapsed() time.Duration {
	return time.Since(s.start).Truncate(time.Second)
}

func FormatDuration(d time.Duration) string {
	m := int(d.Minutes())
	sec := int(d.Seconds()) % secondsPerMinute
	if m > 0 {
		return fmt.Sprintf("%dm %02ds", m, sec)
	}
	return fmt.Sprintf("%ds", sec)
}

type fdWriter interface {
	Fd() uintptr
}

func IsTerminal(w io.Writer) bool {
	if f, ok := w.(fdWriter); ok {
		return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}
	return false
}
