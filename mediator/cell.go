package mediator

const (
	empty  = 0
	banana = 1
	wall   = 2
)

type Cell int

func (c Cell) Empty() bool {
	return c == empty
}

func (c Cell) Banana() bool {
	return c == banana
}

func (c Cell) Wall() bool {
	return c == wall
}

func (c Cell) Player() bool {
	return c > wall
}
