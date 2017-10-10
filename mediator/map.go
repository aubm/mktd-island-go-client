package mediator

import "errors"

var (
	ErrOutOfRange = errors.New("out of range")
)

type Map [][]Cell

func (m Map) Cell(x, y int) (c Cell, err error) {
	if x < 0 || y < 0 {
		return c, ErrOutOfRange
	}

	if y >= len(m) {
		return c, ErrOutOfRange
	}

	r := m[y]

	if x >= len(r) {
		return c, ErrOutOfRange
	}

	return m[y][x], nil
}
