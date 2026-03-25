package platform

import "github.com/tomas-mraz/input"

const (
	widthWindow  = 740
	heightWindow = 360
)

// Backend je abstraktní platformní vrstva.
// Konzument inputu používá stejné API bez ohledu na backend (GLFW/NDK).
type Backend interface {
	TimeSeconds() float64
	PumpEvents() bool
	Close()
}

type Config struct {
	Width  int
	Height int
	Title  string
}

func DefaultConfig() Config {
	return Config{
		Width:  widthWindow,
		Height: heightWindow,
		Title:  "go-input events",
	}
}

func New(in *input.Input, cfg Config) (Backend, error) {
	if cfg.Width <= 0 {
		cfg.Width = widthWindow
	}
	if cfg.Height <= 0 {
		cfg.Height = heightWindow
	}
	if cfg.Title == "" {
		cfg.Title = "go-input events"
	}
	return newBackend(in, cfg)
}
