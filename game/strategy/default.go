package strategy

import (
	"mktd5/mktd-island/client/log"
	"mktd5/mktd-island/client/mediator"
)

type DefaultMoveStrategy struct {
	Logger log.LoggerInterface `inject:""`
}

// DecideWhereToGo is invoked when it is the player's turn to make a move.
// Given a helper object that can provide useful context, it has to decide (in a limited
// time frame) what will be the next move for the player.
// The function must return one of the following directions:
//   - mediator.North
//   - mediator.East
//   - mediator.South
//   - mediator.West
//   - mediator.None
// It can also return an error. If an error is returned, the player will make no move
// for the current turn and the error will be logged.
func (d *DefaultMoveStrategy) DecideWhereToGo(helper Helper) (mediator.Direction, error) {
	d.Logger.Info("let's go!", nil)
	return mediator.None, nil
}
