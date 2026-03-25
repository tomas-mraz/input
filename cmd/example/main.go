package main

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/tomas-mraz/input"
	"github.com/tomas-mraz/input/platform"
)

type eventTracker struct {
	padConnected map[int]bool
	padAxes      map[int]map[input.GamepadAxis]float32
}

func newEventTracker() *eventTracker {
	return &eventTracker{
		padConnected: map[int]bool{},
		padAxes:      map[int]map[input.GamepadAxis]float32{},
	}
}

func hasInputEvent(in *input.Input, tr *eventTracker) bool {
	for _, s := range in.Keys {
		if s.JustPressed || s.JustReleased {
			return true
		}
	}

	for _, s := range in.Mouse.Buttons {
		if s.JustPressed || s.JustReleased {
			return true
		}
	}
	if in.Mouse.DX != 0 || in.Mouse.DY != 0 || in.Mouse.ScrollX != 0 || in.Mouse.ScrollY != 0 {
		return true
	}

	const axisEpsilon = 0.01
	axes := [...]input.GamepadAxis{
		input.AxisLeftX, input.AxisLeftY,
		input.AxisRightX, input.AxisRightY,
		input.AxisLT, input.AxisRT,
	}

	for idx, gp := range in.Gamepads {
		prevConn, knownConn := tr.padConnected[idx]
		if !knownConn {
			tr.padConnected[idx] = gp.Connected
		} else if prevConn != gp.Connected {
			tr.padConnected[idx] = gp.Connected
			return true
		}

		for _, b := range gp.Buttons {
			if b.JustPressed || b.JustReleased {
				return true
			}
		}

		if tr.padAxes[idx] == nil {
			tr.padAxes[idx] = map[input.GamepadAxis]float32{}
		}
		for _, a := range axes {
			cur := gp.Axis(a)
			prev, known := tr.padAxes[idx][a]
			tr.padAxes[idx][a] = cur
			if known && math.Abs(float64(cur-prev)) > axisEpsilon {
				return true
			}
		}
	}

	return false
}

func keyName(k input.Key) string {
	switch k {
	case input.KeyW:
		return "W"
	case input.KeyA:
		return "A"
	case input.KeyS:
		return "S"
	case input.KeyD:
		return "D"
	case input.KeySpace:
		return "Space"
	case input.KeyShift:
		return "Shift"
	case input.KeyCtrl:
		return "Ctrl"
	case input.KeyEscape:
		return "Escape"
	case input.KeyLeft:
		return "Left"
	case input.KeyRight:
		return "Right"
	case input.KeyUp:
		return "Up"
	case input.KeyDown:
		return "Down"
	default:
		return fmt.Sprintf("Key(%d)", int(k))
	}
}

func stateToken(down, jp, jr bool, held float64) string {
	if jp {
		return "P"
	}
	if jr {
		return fmt.Sprintf("R%.2f", held)
	}
	if down {
		return fmt.Sprintf("D%.2f", held)
	}
	return "-"
}

func keySummary(in *input.Input) string {
	keys := make([]int, 0, len(in.Keys))
	for k, s := range in.Keys {
		if s.Down || s.JustPressed || s.JustReleased {
			keys = append(keys, int(k))
		}
	}
	if len(keys) == 0 {
		return "-"
	}
	sort.Ints(keys)

	parts := make([]string, 0, len(keys))
	for _, kv := range keys {
		k := input.Key(kv)
		s := in.Key(k)
		held := 0.0
		if s.Down {
			held = in.KeyHeldDuration(k)
		} else if s.JustReleased {
			held = s.ReleasedAt - s.PressedAt
		}
		parts = append(parts, fmt.Sprintf("%s:%s", keyName(k), stateToken(s.Down, s.JustPressed, s.JustReleased, held)))
	}
	return strings.Join(parts, ", ")
}

func mouseButtonName(b input.MouseButton) string {
	switch b {
	case input.MouseLeft:
		return "L"
	case input.MouseRight:
		return "R"
	case input.MouseMiddle:
		return "M"
	default:
		return fmt.Sprintf("B%d", int(b))
	}
}

func mouseSummary(in *input.Input) string {
	buttons := make([]int, 0, len(in.Mouse.Buttons))
	for b, s := range in.Mouse.Buttons {
		if s.Down || s.JustPressed || s.JustReleased {
			buttons = append(buttons, int(b))
		}
	}
	if len(buttons) == 0 {
		return "-"
	}
	sort.Ints(buttons)

	parts := make([]string, 0, len(buttons))
	for _, bv := range buttons {
		b := input.MouseButton(bv)
		s := in.Mouse.Buttons[b]
		held := 0.0
		if s.Down {
			held = in.MouseHeldDuration(b)
		} else if s.JustReleased {
			held = s.ReleasedAt - s.PressedAt
		}
		parts = append(parts, fmt.Sprintf("%s:%s", mouseButtonName(b), stateToken(s.Down, s.JustPressed, s.JustReleased, held)))
	}

	return strings.Join(parts, ", ")
}

func padSummary(in *input.Input) string {
	gp := in.Gamepad(0)
	if gp == nil {
		return "-"
	}
	if !gp.Connected {
		return "off"
	}
	a := gp.Buttons[input.PadA]
	aToken := "-"
	if a != nil {
		held := 0.0
		if a.Down {
			held = in.GamepadButtonHeldDuration(0, input.PadA)
		} else if a.JustReleased {
			held = a.ReleasedAt - a.PressedAt
		}
		aToken = stateToken(a.Down, a.JustPressed, a.JustReleased, held)
	}
	return fmt.Sprintf("on A:%s lx:%.2f", aToken, gp.Axis(input.AxisLeftX))
}

func logInputs(now float64, in *input.Input) {
	fmt.Printf(
		"t=%4.2f k=[%s] m=(%.0f,%.0f d=%.0f,%.0f b=[%s] s=%.0f,%.0f) p0=[%s]\n",
		now,
		keySummary(in),
		in.Mouse.X, in.Mouse.Y, in.Mouse.DX, in.Mouse.DY,
		mouseSummary(in),
		in.Mouse.ScrollX, in.Mouse.ScrollY,
		padSummary(in),
	)
}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	fmt.Println("start")

	in := input.New()
	backend, err := platform.New(in, platform.DefaultConfig())
	if err != nil {
		fmt.Printf("platform backend init failed: %v\n", err)
		return
	}
	defer backend.Close()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	tracker := newEventTracker()

	running := true
	for running {
		<-ticker.C

		now := backend.TimeSeconds() // monotonic time in seconds
		in.Tick(now)

		// platform backend zde zavolá in.KeyDown/Up, in.MouseMove, ...
		running = backend.PumpEvents()

		if hasInputEvent(in, tracker) {
			logInputs(now, in)
		}
	}

	fmt.Println("end")
}
