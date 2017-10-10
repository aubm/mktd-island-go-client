package mediator

var (
	North Direction = "N"
	East  Direction = "E"
	South Direction = "S"
	West  Direction = "W"
	None  Direction = "O"
)

type Direction string

func (d Direction) North() bool {
	return d == North
}

func (d Direction) East() bool {
	return d == East
}

func (d Direction) South() bool {
	return d == South
}

func (d Direction) West() bool {
	return d == West
}

func (d Direction) None() bool {
	return d == None
}
