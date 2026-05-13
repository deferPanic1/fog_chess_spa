package service

import (
	"testing"

	chessengine "github.com/Gilf4/fog_chess/internal/chessEngine"
)

func TestDefaultPromotionMoveAppliesOnlyToPawnOnLastRank(t *testing.T) {
	board, err := chessengine.ParseFEN("7k/8/8/8/8/8/6p1/7R b - - 0 1")
	if err != nil {
		t.Fatalf("ParseFEN returned error: %v", err)
	}

	move, err := chessengine.ParseMove("g2", "h1", "")
	if err != nil {
		t.Fatalf("ParseMove returned error: %v", err)
	}

	move = defaultPromotionMove(*board, move)
	if move.Promotion != chessengine.Queen {
		t.Fatalf("expected queen promotion, got %v", move.Promotion)
	}
}

func TestDefaultPromotionMoveDoesNotAffectNonPawn(t *testing.T) {
	board, err := chessengine.ParseFEN("7k/8/8/8/8/8/6Q1/K7 w - - 0 1")
	if err != nil {
		t.Fatalf("ParseFEN returned error: %v", err)
	}

	move, err := chessengine.ParseMove("g2", "h1", "")
	if err != nil {
		t.Fatalf("ParseMove returned error: %v", err)
	}

	move = defaultPromotionMove(*board, move)
	if move.Promotion != chessengine.Empty {
		t.Fatalf("expected no promotion, got %v", move.Promotion)
	}
}
