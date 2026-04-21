package input

import "math"

type Key int
type MouseButton int
type GamepadButton int
type GamepadAxis int

// Doporučení: drž si vlastní enumy.
// Začni minimem, rozšíříš později.
const (
	KeyUnknown Key = iota
	KeyW
	KeyA
	KeyS
	KeyD
	KeySpace
	KeyEnter
	KeyShift
	KeyCtrl
	KeyEscape
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
)

const (
	MouseLeft MouseButton = iota
	MouseRight
	MouseMiddle
)

const (
	PadA GamepadButton = iota
	PadB
	PadX
	PadY
	PadLB
	PadRB
	PadBack
	PadStart
	PadLStick
	PadRStick
	PadDpadUp
	PadDpadDown
	PadDpadLeft
	PadDpadRight
)

const (
	AxisLeftX GamepadAxis = iota
	AxisLeftY
	AxisRightX
	AxisRightY
	AxisLT
	AxisRT
)

type ButtonState struct {
	Down         bool
	JustPressed  bool
	JustReleased bool
	PressedAt    float64
	ReleasedAt   float64
}

type GamepadState struct {
	Connected bool

	Buttons map[GamepadButton]*ButtonState
	Axes    map[GamepadAxis]float32

	Deadzone float32 // např. 0.15
}

type Input struct {
	now  float64
	prev float64

	Keys  map[Key]*ButtonState
	Mouse struct {
		X, Y         float64
		DX, DY       float64
		ScrollX      float64
		ScrollY      float64
		Buttons      map[MouseButton]*ButtonState
		HasPosition  bool
		prevX, prevY float64
	}
	Gamepads map[int]*GamepadState // index 0..N
}

func New() *Input {
	in := &Input{
		Keys:     map[Key]*ButtonState{},
		Gamepads: map[int]*GamepadState{},
	}
	in.Mouse.Buttons = map[MouseButton]*ButtonState{}
	return in
}

// Tick zavolej 1× za frame (před zpracováním hry).
// now = monotonic time v sekundách (např. glfw.GetTime() nebo clock_gettime na Androidu).
func (in *Input) Tick(now float64) {
	in.prev = in.now
	in.now = now

	// reset edge flags (JustPressed/JustReleased) a per-frame delty
	for _, s := range in.Keys {
		s.JustPressed = false
		s.JustReleased = false
	}
	for _, s := range in.Mouse.Buttons {
		s.JustPressed = false
		s.JustReleased = false
	}
	for _, gp := range in.Gamepads {
		for _, b := range gp.Buttons {
			b.JustPressed = false
			b.JustReleased = false
		}
	}

	in.Mouse.DX, in.Mouse.DY = 0, 0
	in.Mouse.ScrollX, in.Mouse.ScrollY = 0, 0
}

func (in *Input) DT() float64 {
	dt := in.now - in.prev
	if dt < 0 {
		return 0
	}
	return dt
}

func (in *Input) ensureKey(k Key) *ButtonState {
	s := in.Keys[k]
	if s == nil {
		s = &ButtonState{}
		in.Keys[k] = s
	}
	return s
}

func (in *Input) ensureMouseButton(b MouseButton) *ButtonState {
	s := in.Mouse.Buttons[b]
	if s == nil {
		s = &ButtonState{}
		in.Mouse.Buttons[b] = s
	}
	return s
}

func (in *Input) ensureGamepad(idx int) *GamepadState {
	gp := in.Gamepads[idx]
	if gp == nil {
		gp = &GamepadState{
			Buttons:  map[GamepadButton]*ButtonState{},
			Axes:     map[GamepadAxis]float32{},
			Deadzone: 0.15,
		}
		in.Gamepads[idx] = gp
	}
	return gp
}

func (gp *GamepadState) ensureButton(b GamepadButton) *ButtonState {
	s := gp.Buttons[b]
	if s == nil {
		s = &ButtonState{}
		gp.Buttons[b] = s
	}
	return s
}

// ---------- Raw event API (platform layer tohle volá) ----------

func (in *Input) KeyDown(k Key) {
	if k == KeyUnknown {
		return
	}
	s := in.ensureKey(k)
	if s.Down {
		return
	}
	s.Down = true
	s.JustPressed = true
	s.PressedAt = in.now
}

func (in *Input) KeyUp(k Key) {
	if k == KeyUnknown {
		return
	}
	s := in.ensureKey(k)
	if !s.Down {
		return
	}
	s.Down = false
	s.JustReleased = true
	s.ReleasedAt = in.now
}

func (in *Input) MouseMove(x, y float64) {
	if !in.Mouse.HasPosition {
		in.Mouse.HasPosition = true
		in.Mouse.prevX, in.Mouse.prevY = x, y
		in.Mouse.X, in.Mouse.Y = x, y
		return
	}
	in.Mouse.DX += x - in.Mouse.prevX
	in.Mouse.DY += y - in.Mouse.prevY
	in.Mouse.prevX, in.Mouse.prevY = x, y
	in.Mouse.X, in.Mouse.Y = x, y
}

func (in *Input) MouseButtonDown(b MouseButton) {
	s := in.ensureMouseButton(b)
	if s.Down {
		return
	}
	s.Down = true
	s.JustPressed = true
	s.PressedAt = in.now
}

func (in *Input) MouseButtonUp(b MouseButton) {
	s := in.ensureMouseButton(b)
	if !s.Down {
		return
	}
	s.Down = false
	s.JustReleased = true
	s.ReleasedAt = in.now
}

func (in *Input) MouseScroll(dx, dy float64) {
	in.Mouse.ScrollX += dx
	in.Mouse.ScrollY += dy
}

func (in *Input) GamepadConnected(idx int, connected bool) {
	gp := in.ensureGamepad(idx)
	gp.Connected = connected
}

func (in *Input) GamepadButtonDown(idx int, b GamepadButton) {
	gp := in.ensureGamepad(idx)
	s := gp.ensureButton(b)
	if s.Down {
		return
	}
	s.Down = true
	s.JustPressed = true
	s.PressedAt = in.now
}

func (in *Input) GamepadButtonUp(idx int, b GamepadButton) {
	gp := in.ensureGamepad(idx)
	s := gp.ensureButton(b)
	if !s.Down {
		return
	}
	s.Down = false
	s.JustReleased = true
	s.ReleasedAt = in.now
}

func (in *Input) GamepadAxis(idx int, a GamepadAxis, v float32) {
	gp := in.ensureGamepad(idx)
	// deadzone pro stick osy (triggerům často deadzone nedávej, ale jednoduché řešení necháme univerzální)
	if gp.Deadzone > 0 && (a == AxisLeftX || a == AxisLeftY || a == AxisRightX || a == AxisRightY) {
		if float32(math.Abs(float64(v))) < gp.Deadzone {
			v = 0
		}
	}
	gp.Axes[a] = v
}

// ---------- Gameplay-friendly query API ----------

func (in *Input) Key(k Key) ButtonState {
	s := in.Keys[k]
	if s == nil {
		return ButtonState{}
	}
	return *s
}

func (in *Input) KeyDownQ(k Key) bool        { return in.Keys[k] != nil && in.Keys[k].Down }
func (in *Input) KeyJustPressed(k Key) bool  { return in.Keys[k] != nil && in.Keys[k].JustPressed }
func (in *Input) KeyJustReleased(k Key) bool { return in.Keys[k] != nil && in.Keys[k].JustReleased }

func (in *Input) KeyHeldDuration(k Key) float64 {
	s := in.Keys[k]
	if s == nil || !s.Down {
		return 0
	}
	return in.now - s.PressedAt
}

func (in *Input) MouseDown(b MouseButton) bool {
	return in.Mouse.Buttons[b] != nil && in.Mouse.Buttons[b].Down
}
func (in *Input) MouseJustPressed(b MouseButton) bool {
	return in.Mouse.Buttons[b] != nil && in.Mouse.Buttons[b].JustPressed
}
func (in *Input) MouseJustReleased(b MouseButton) bool {
	return in.Mouse.Buttons[b] != nil && in.Mouse.Buttons[b].JustReleased
}
func (in *Input) MouseHeldDuration(b MouseButton) float64 {
	s := in.Mouse.Buttons[b]
	if s == nil || !s.Down {
		return 0
	}
	return in.now - s.PressedAt
}

func (in *Input) Gamepad(idx int) *GamepadState {
	return in.Gamepads[idx]
}

func (in *Input) GamepadButtonHeldDuration(idx int, b GamepadButton) float64 {
	gp := in.Gamepads[idx]
	if gp == nil {
		return 0
	}
	return gp.HeldDuration(in.now, b)
}

func (gp *GamepadState) ButtonDown(b GamepadButton) bool {
	return gp.Buttons[b] != nil && gp.Buttons[b].Down
}
func (gp *GamepadState) ButtonJustPressed(b GamepadButton) bool {
	return gp.Buttons[b] != nil && gp.Buttons[b].JustPressed
}
func (gp *GamepadState) ButtonJustReleased(b GamepadButton) bool {
	return gp.Buttons[b] != nil && gp.Buttons[b].JustReleased
}
func (gp *GamepadState) HeldDuration(now float64, b GamepadButton) float64 {
	s := gp.Buttons[b]
	if s == nil || !s.Down {
		return 0
	}
	return now - s.PressedAt
}
func (gp *GamepadState) Axis(a GamepadAxis) float32 { return gp.Axes[a] }
