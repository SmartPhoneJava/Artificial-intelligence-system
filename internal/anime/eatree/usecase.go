package eatree

import "io"

type Visualizer interface {
	Render(w ...io.Writer) error
}
