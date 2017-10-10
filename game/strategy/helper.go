package strategy

import "mktd5/mktd-island/client/mediator"

type Helper interface {
	// GameState returns the most recent snapshot of the game state
	// which contains a multi-dimensional array of cells and the list of players.
	GameState() mediator.State

	// Me returns true if the provided player is the current player.
	IsMe(playerId mediator.Cell) bool
}
