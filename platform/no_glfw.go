//go:build !android && !glfw

package platform

import (
	"fmt"

	"github.com/tomas-mraz/input"
)

func newBackend(_ *input.Input, _ Config) (Backend, error) {
	return nil, fmt.Errorf("desktop backend requires build tag 'glfw' and GLFW development libraries")
}
