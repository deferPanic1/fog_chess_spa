package chessengine

type GameState struct {
	Board Board
	Moves []Move
}

type Move struct {
	From      int
	To        int
	Promotion PieceType
}
