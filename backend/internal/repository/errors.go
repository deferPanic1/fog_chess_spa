package repository

import "errors"

var (
	//user errors
	ErrNotFound       = errors.New("user not found")
	ErrEmailExists    = errors.New("email already exists")
	ErrUsernameExists = errors.New("username already exists")
	ErrUserInUse      = errors.New("user is referenced by other records")

	//lobby errors
	ErrInvalidTimeControl  = errors.New("invalid time control")
	ErrUserAlreadyInLobby  = errors.New("user already in lobby")
	ErrLobbyClosed         = errors.New("lobby is closed")
	ErrLobbyFull           = errors.New("lobby is full")
	ErrLobbyNotParticipant = errors.New("user is not a lobby participant")

	// match errors
	ErrUserAlreadyInMatch = errors.New("user already in active match")
)
