package game

import (
	"mktd5/mktd-island/client/game/strategy"
	"mktd5/mktd-island/client/mediator"
)

type moveStrategy interface {
	DecideWhereToGo(helper strategy.Helper) (mediator.Direction, error)
}

type strategyHelper struct {
	state    mediator.State
	playerID int
}

func (h strategyHelper) GameState() mediator.State {
	return h.state
}

func (h strategyHelper) IsMe(playerId mediator.Cell) bool {
	return int(playerId) == h.playerID
}
