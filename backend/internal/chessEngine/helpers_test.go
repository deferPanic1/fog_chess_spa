package chessengine

import "testing"

func TestParseMoveDoesNotAssumePromotionWithoutExplicitPiece(t *testing.T) {
	move, err := ParseMove("g2", "h1", "")
	if err != nil {
		t.Fatalf("ParseMove returned error: %v", err)
	}

	if move.Promotion != Empty {
		t.Fatalf("expected no implicit promotion in parsed move, got %v", move.Promotion)
	}
}

func TestBlackPawnCanCaptureOnPromotionSquareWithoutExplicitPromotion(t *testing.T) {
	move, err := ParseMove("g2", "h1", "")
	if err != nil {
		t.Fatalf("ParseMove returned error: %v", err)
	}

	if move.Promotion != Empty {
		t.Fatalf("expected no implicit promotion in parsed move, got %v", move.Promotion)
	}
}

func TestQueenCanMoveToBackRankWithoutPromotion(t *testing.T) {
	game, err := NewGameFromFEN("7k/8/8/8/8/8/6Q1/K7 w - - 0 1")
	if err != nil {
		t.Fatalf("NewGameFromFEN returned error: %v", err)
	}

	move, err := ParseMove("g2", "h1", "")
	if err != nil {
		t.Fatalf("ParseMove returned error: %v", err)
	}

	_, err = game.DoMove(move)
	if err != nil {
		t.Fatalf("DoMove returned error: %v", err)
	}

	idx, err := ConverToIndex("h1")
	if err != nil {
		t.Fatalf("ConverToIndex returned error: %v", err)
	}

	piece := game.Board.Squares[idx]
	if piece.Type != Queen || piece.Color != White {
		t.Fatalf("expected white queen on h1, got %+v", piece)
	}
}
