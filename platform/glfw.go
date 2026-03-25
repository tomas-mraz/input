//go:build !android && glfw

package platform

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/tomas-mraz/input"
)

type glfwBackend struct {
	window *glfw.Window
	in     *input.Input

	connected map[int]bool
	buttons   map[int]map[input.GamepadButton]bool
}

func newBackend(in *input.Input, cfg Config) (Backend, error) {
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	glfw.DefaultWindowHints()
	glfw.WindowHint(glfw.Visible, glfw.True)      // ensure window is mapped on create
	glfw.WindowHint(glfw.Resizable, glfw.False)   // stable tiny capture window
	glfw.WindowHint(glfw.Focused, glfw.True)      // request focus on create
	glfw.WindowHint(glfw.AutoIconify, glfw.False) // keep active in windowed mode

	window, err := glfw.CreateWindow(cfg.Width, cfg.Height, cfg.Title, nil, nil)
	if err != nil {
		glfw.Terminate()
		return nil, err
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(1)
	window.Show()
	window.Focus()

	b := &glfwBackend{
		window:    window,
		in:        in,
		connected: map[int]bool{},
		buttons:   map[int]map[input.GamepadButton]bool{},
	}
	b.hookWindowEvents()
	return b, nil
}

func (b *glfwBackend) TimeSeconds() float64 {
	return glfw.GetTime()
}

func (b *glfwBackend) PumpEvents() bool {
	glfw.PollEvents()
	b.window.SwapBuffers()
	b.pollGamepads()
	return !b.window.ShouldClose()
}

func (b *glfwBackend) Close() {
	if b.window != nil {
		b.window.Destroy()
		b.window = nil
	}
	glfw.Terminate()
}

func mapKey(k glfw.Key) input.Key {
	switch k {
	case glfw.KeyW:
		return input.KeyW
	case glfw.KeyA:
		return input.KeyA
	case glfw.KeyS:
		return input.KeyS
	case glfw.KeyD:
		return input.KeyD
	case glfw.KeySpace:
		return input.KeySpace
	case glfw.KeyLeftShift:
		return input.KeyShift
	case glfw.KeyLeftControl:
		return input.KeyCtrl
	case glfw.KeyEscape:
		return input.KeyEscape
	case glfw.KeyLeft:
		return input.KeyLeft
	case glfw.KeyRight:
		return input.KeyRight
	case glfw.KeyUp:
		return input.KeyUp
	case glfw.KeyDown:
		return input.KeyDown
	default:
		if k == glfw.KeyUnknown {
			return input.KeyUnknown
		}
		// Preserve raw GLFW key code for keys outside our named enum.
		return input.Key(k)
	}
}

func mapMouseButton(b glfw.MouseButton) (input.MouseButton, bool) {
	switch b {
	case glfw.MouseButtonLeft:
		return input.MouseLeft, true
	case glfw.MouseButtonRight:
		return input.MouseRight, true
	case glfw.MouseButtonMiddle:
		return input.MouseMiddle, true
	default:
		return 0, false
	}
}

func (b *glfwBackend) hookWindowEvents() {
	w := b.window
	in := b.in

	w.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		k := mapKey(key)
		switch action {
		case glfw.Press:
			in.KeyDown(k)
		case glfw.Release:
			in.KeyUp(k)
		}
	})

	w.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		in.MouseMove(x, y)
	})

	w.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, _ glfw.ModifierKey) {
		mb, ok := mapMouseButton(button)
		if !ok {
			return
		}
		switch action {
		case glfw.Press:
			in.MouseButtonDown(mb)
		case glfw.Release:
			in.MouseButtonUp(mb)
		}
	})

	w.SetScrollCallback(func(_ *glfw.Window, dx, dy float64) {
		in.MouseScroll(dx, dy)
	})
}

func mapPadButton(gb glfw.GamepadButton) (input.GamepadButton, bool) {
	switch gb {
	case glfw.ButtonA:
		return input.PadA, true
	case glfw.ButtonB:
		return input.PadB, true
	case glfw.ButtonX:
		return input.PadX, true
	case glfw.ButtonY:
		return input.PadY, true
	case glfw.ButtonLeftBumper:
		return input.PadLB, true
	case glfw.ButtonRightBumper:
		return input.PadRB, true
	case glfw.ButtonBack:
		return input.PadBack, true
	case glfw.ButtonStart:
		return input.PadStart, true
	case glfw.ButtonLeftThumb:
		return input.PadLStick, true
	case glfw.ButtonRightThumb:
		return input.PadRStick, true
	case glfw.ButtonDpadUp:
		return input.PadDpadUp, true
	case glfw.ButtonDpadDown:
		return input.PadDpadDown, true
	case glfw.ButtonDpadLeft:
		return input.PadDpadLeft, true
	case glfw.ButtonDpadRight:
		return input.PadDpadRight, true
	default:
		return 0, false
	}
}

func (b *glfwBackend) syncGamepadButton(idx int, btn input.GamepadButton, down bool) {
	if b.buttons[idx] == nil {
		b.buttons[idx] = map[input.GamepadButton]bool{}
	}
	wasDown := b.buttons[idx][btn]
	if down == wasDown {
		return
	}
	b.buttons[idx][btn] = down
	if down {
		b.in.GamepadButtonDown(idx, btn)
	} else {
		b.in.GamepadButtonUp(idx, btn)
	}
}

func (b *glfwBackend) pollGamepads() {
	for joy := glfw.Joystick1; joy <= glfw.JoystickLast; joy++ {
		idx := int(joy - glfw.Joystick1)
		present := joy.Present() && joy.IsGamepad()

		if b.connected[idx] != present {
			if !present {
				for btn, wasDown := range b.buttons[idx] {
					if wasDown {
						b.in.GamepadButtonUp(idx, btn)
					}
				}
				delete(b.buttons, idx)
			}
			b.connected[idx] = present
			b.in.GamepadConnected(idx, present)
		}
		if !present {
			continue
		}

		state := joy.GetGamepadState()
		if state == nil {
			continue
		}

		// Button states are polled, so we synthesize down/up transitions.
		for gb := glfw.ButtonA; gb <= glfw.ButtonDpadLeft; gb++ {
			btn, ok := mapPadButton(gb)
			if !ok {
				continue
			}
			down := state.Buttons[gb] == glfw.Press
			b.syncGamepadButton(idx, btn, down)
		}

		b.in.GamepadAxis(idx, input.AxisLeftX, state.Axes[glfw.AxisLeftX])
		b.in.GamepadAxis(idx, input.AxisLeftY, state.Axes[glfw.AxisLeftY])
		b.in.GamepadAxis(idx, input.AxisRightX, state.Axes[glfw.AxisRightX])
		b.in.GamepadAxis(idx, input.AxisRightY, state.Axes[glfw.AxisRightY])
		b.in.GamepadAxis(idx, input.AxisLT, state.Axes[glfw.AxisLeftTrigger])
		b.in.GamepadAxis(idx, input.AxisRT, state.Axes[glfw.AxisRightTrigger])
	}
}
