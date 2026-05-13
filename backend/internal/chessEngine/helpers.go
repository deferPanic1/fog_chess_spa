package chessengine

import (
	"fmt"
	"sort"
	"strings"
)

const InitialFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func ParseFEN(fen string) (*Board, error) {
	return parseFen(fen)
}

func NewGameFromFEN(fen string) (*GameState, error) {
	board, err := parseFen(fen)
	if err != nil {
		return nil, err
	}

	return &GameState{Board: *board}, nil
}

func (g *GameState) FEN() (string, error) {
	return g.Board.formatFen()
}

func ColorName(color Color) string {
	switch color {
	case White:
		return "white"
	case Black:
		return "black"
	default:
		return ""
	}
}

func ParseMove(from, to, promotion string) (Move, error) {
	fromIdx, err := ConverToIndex(from)
	if err != nil {
		return Move{}, fmt.Errorf("invalid from square: %w", err)
	}

	toIdx, err := ConverToIndex(to)
	if err != nil {
		return Move{}, fmt.Errorf("invalid to square: %w", err)
	}

	move := Move{From: fromIdx, To: toIdx}
	if promotion == "" {
		return move, nil
	}

	piece, err := promotionPiece(promotion)
	if err != nil {
		return Move{}, err
	}
	move.Promotion = piece

	return move, nil
}

func promotionPiece(value string) (PieceType, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "q", "queen":
		return Queen, nil
	case "r", "rook":
		return Rook, nil
	case "b", "bishop":
		return Bishop, nil
	case "n", "k", "knight":
		return Knight, nil
	default:
		return Empty, fmt.Errorf("unsupported promotion piece %q", value)
	}
}

func (b *Board) LegalMovesFor(color Color) map[string][]string {
	legal := make(map[string][]string)
	if b.SideToMove != color {
		return legal
	}

	for idx, piece := range b.Squares {
		if piece.Color != color {
			continue
		}

		moves := b.pseudoLegalMoves(idx)
		if len(moves) == 0 {
			continue
		}

		from := IndexToCoord(idx)
		for _, move := range moves {
			legal[from] = append(legal[from], IndexToCoord(move.To))
		}
		sort.Strings(legal[from])
	}

	return legal
}

func (b *Board) FoggedSquaresFor(color Color) []string {
	visible := b.VisibleSquares(color)
	squares := make([]string, 0, 64-len(visible))
	for idx := range b.Squares {
		if visible[idx] {
			continue
		}
		squares = append(squares, IndexToCoord(idx))
	}
	sort.Strings(squares)
	return squares
}

func (b *Board) DisplayFENForPlayer(color Color) (string, error) {
	playerBoard := b.BoardForPlayer(color)
	for idx, piece := range playerBoard.Squares {
		if piece.Type != Invisible {
			continue
		}
		playerBoard.Squares[idx] = Piece{Type: Empty, Color: NoColor}
	}

	if color == White {
		playerBoard.BlackKingside = false
		playerBoard.BlackQueenside = false
	} else {
		playerBoard.WhiteKingside = false
		playerBoard.WhiteQueenside = false
	}

	return playerBoard.formatFen()
}
