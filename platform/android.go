//go:build android

package platform

import (
	"time"
	
	"github.com/tomas-mraz/input"
)

// ndkBackend je Android větev platformní abstrakce.
// Zde se budou z AInputEvent/AKeyEvent/AMotionEvent mapovat raw eventy do input.Input.
type ndkBackend struct {
	start time.Time
}

func newBackend(_ *input.Input, _ Config) (Backend, error) {
	return &ndkBackend{start: time.Now()}, nil
}

func (b *ndkBackend) TimeSeconds() float64 {
	return time.Since(b.start).Seconds()
}

func (b *ndkBackend) PumpEvents() bool {
	// TODO: napojit NDK event queue a volat in.KeyDown/Up, MouseMove, Gamepad...
	return true
}

func (b *ndkBackend) Close() {}
