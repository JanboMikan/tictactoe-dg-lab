package game

import "errors"

var (
	// Room errors
	ErrRoomNotFound   = errors.New("room not found")
	ErrRoomFull       = errors.New("room is full")
	ErrPlayerNotFound = errors.New("player not found")

	// Game errors
	ErrGameOver       = errors.New("game is already over")
	ErrNotYourTurn    = errors.New("not your turn")
	ErrInvalidMove    = errors.New("invalid move")
	ErrPositionOccupied = errors.New("position already occupied")

	// Config errors
	ErrInvalidConfig  = errors.New("invalid player configuration")

	// Permission errors
	ErrNotWinner      = errors.New("only winner can punish")
)
