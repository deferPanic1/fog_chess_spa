package chessengine

import "errors"

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func rankOf(idx int) int { return idx / 8 }
func fileOf(idx int) int { return idx % 8 }

func opponent(c Color) Color {
	if c == White {
		return Black
	}
	return White
}

func (b *Board) VisibleSquares(color Color) map[int]bool {
	visible := make(map[int]bool, 64)
	for i, p := range b.Squares {
		if p.Color != color {
			continue
		}
		visible[i] = true
		for _, sq := range b.visibilitySquares(i) {
			visible[sq] = true
		}
	}
	return visible
}

// visibilitySquares — клетки которые фигура видит.
func (b *Board) visibilitySquares(from int) []int {
	p := b.Squares[from]
	switch p.Type {
	case Pawn:
		return b.pawnVisibility(from, p.Color)
	case Knight:
		r, f := rankOf(from), fileOf(from)
		deltas := [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
		var res []int
		for _, d := range deltas {
			nr, nf := r+d[0], f+d[1]
			if nr >= 0 && nr < 8 && nf >= 0 && nf < 8 {
				res = append(res, nr*8+nf)
			}
		}
		return res
	case Bishop:
		return b.slideVisibility(from, [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}})
	case Rook:
		return b.slideVisibility(from, [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}})
	case Queen:
		return b.slideVisibility(from, [][2]int{
			{1, 0}, {-1, 0}, {0, 1}, {0, -1},
			{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
		})
	case King:
		r, f := rankOf(from), fileOf(from)
		var res []int
		for dr := -1; dr <= 1; dr++ {
			for df := -1; df <= 1; df++ {
				if dr == 0 && df == 0 {
					continue
				}
				nr, nf := r+dr, f+df
				if nr >= 0 && nr < 8 && nf >= 0 && nf < 8 {
					res = append(res, nr*8+nf)
				}
			}
		}
		return res
	}
	return nil
}

// pawnVisibility — пешка видит клетку впереди И обе диагонали всегда
func (b *Board) pawnVisibility(from int, color Color) []int {
	r, f := rankOf(from), fileOf(from)
	dir := -1
	startRank := 6
	if color == Black {
		dir = 1
		startRank = 1
	}

	nr := r + dir
	if nr < 0 || nr >= 8 {
		return nil
	}

	res := []int{nr*8 + f}

	for _, df := range []int{-1, 1} {
		nf := f + df
		if nf >= 0 && nf < 8 {
			res = append(res, nr*8+nf)
		}
	}

	if r == startRank && b.Squares[nr*8+f].Type == Empty {
		nr2 := r + dir*2
		res = append(res, nr2*8+f)
	}

	return res
}

// slideVisibility — диагональный фигуры: видят до первой фигуры включительно
func (b *Board) slideVisibility(from int, dirs [][2]int) []int {
	r, f := rankOf(from), fileOf(from)
	var res []int
	for _, d := range dirs {
		nr, nf := r+d[0], f+d[1]
		for nr >= 0 && nr < 8 && nf >= 0 && nf < 8 {
			sq := nr*8 + nf
			t := b.Squares[sq]
			if t.Type == Invisible {
				break
			}
			res = append(res, sq)
			if t.Type != Empty {
				break // видим фигуру но за неё не смотрим
			}
			nr += d[0]
			nf += d[1]
		}
	}
	return res
}

// BoardForPlayer возвращает доску с туманом для рендеринга конкретному игроку
func (b *Board) BoardForPlayer(color Color) Board {
	visible := b.VisibleSquares(color)
	nb := *b
	for i, p := range nb.Squares {
		if !visible[i] && p.Color != color {
			nb.Squares[i] = Piece{Invisible, NoColor}
		}
	}
	return nb
}

// Псевдо-легальные ходы
func (b *Board) pseudoLegalMoves(from int) []Move {
	p := b.Squares[from]
	switch p.Type {
	case Pawn:
		return b.pawnMoves(from, p.Color)
	case Knight:
		return b.knightMoves(from, p.Color)
	case Bishop:
		return b.slideMoves(from, p.Color, [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}})
	case Rook:
		return b.slideMoves(from, p.Color, [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}})
	case Queen:
		return b.slideMoves(from, p.Color, [][2]int{
			{1, 0}, {-1, 0}, {0, 1}, {0, -1},
			{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
		})
	case King:
		return b.kingMoves(from, p.Color)
	}
	return nil
}

func (b *Board) pawnMoves(from int, color Color) []Move {
	r, f := rankOf(from), fileOf(from)
	dir := -1
	startRank := 6
	if color == Black {
		dir = 1
		startRank = 1
	}

	var moves []Move
	nr := r + dir
	if nr < 0 || nr >= 8 {
		return moves
	}

	sq := nr*8 + f
	if b.Squares[sq].Type == Empty {
		moves = append(moves, pawnWithPromotion(from, sq, color, nr)...)
		if r == startRank {
			nr2 := r + dir*2
			sq2 := nr2*8 + f
			if b.Squares[sq2].Type == Empty {
				moves = append(moves, Move{From: from, To: sq2})
			}
		}
	}

	for _, df := range []int{-1, 1} {
		nf := f + df
		if nf < 0 || nf >= 8 {
			continue
		}
		sq := nr*8 + nf
		t := b.Squares[sq]
		if t.Type != Empty && t.Type != Invisible && t.Color != color {
			moves = append(moves, pawnWithPromotion(from, sq, color, nr)...)
		}
	}

	return moves
}

func pawnWithPromotion(from, to int, color Color, toRank int) []Move {
	promotionRank := 0
	if color == Black {
		promotionRank = 7
	}
	if toRank == promotionRank {
		return []Move{
			{From: from, To: to, Promotion: Queen},
			{From: from, To: to, Promotion: Rook},
			{From: from, To: to, Promotion: Bishop},
			{From: from, To: to, Promotion: Knight},
		}
	}
	return []Move{{From: from, To: to}}
}

func (b *Board) knightMoves(from int, color Color) []Move {
	r, f := rankOf(from), fileOf(from)
	deltas := [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
	var moves []Move
	for _, d := range deltas {
		nr, nf := r+d[0], f+d[1]
		if nr < 0 || nr >= 8 || nf < 0 || nf >= 8 {
			continue
		}
		sq := nr*8 + nf
		t := b.Squares[sq]
		if t.Color != color && t.Type != Invisible {
			moves = append(moves, Move{From: from, To: sq})
		}
	}
	return moves
}

func (b *Board) slideMoves(from int, color Color, dirs [][2]int) []Move {
	r, f := rankOf(from), fileOf(from)
	var moves []Move
	for _, d := range dirs {
		nr, nf := r+d[0], f+d[1]
		for nr >= 0 && nr < 8 && nf >= 0 && nf < 8 {
			sq := nr*8 + nf
			t := b.Squares[sq]
			if t.Type == Invisible {
				break
			}
			if t.Type == Empty {
				moves = append(moves, Move{From: from, To: sq})
			} else {
				if t.Color != color {
					moves = append(moves, Move{From: from, To: sq})
				}
				break
			}
			nr += d[0]
			nf += d[1]
		}
	}
	return moves
}

func (b *Board) kingMoves(from int, color Color) []Move {
	r, f := rankOf(from), fileOf(from)
	var moves []Move

	for dr := -1; dr <= 1; dr++ {
		for df := -1; df <= 1; df++ {
			if dr == 0 && df == 0 {
				continue
			}
			nr, nf := r+dr, f+df
			if nr < 0 || nr >= 8 || nf < 0 || nf >= 8 {
				continue
			}
			sq := nr*8 + nf
			t := b.Squares[sq]
			if t.Color != color && t.Type != Invisible {
				moves = append(moves, Move{From: from, To: sq})
			}
		}
	}

	var kingStart int
	var canKingside, canQueenside bool
	if color == White {
		kingStart = 60
		canKingside = b.WhiteKingside
		canQueenside = b.WhiteQueenside
	} else {
		kingStart = 4
		canKingside = b.BlackKingside
		canQueenside = b.BlackQueenside
	}

	if from == kingStart {
		if canKingside {
			f1, g1 := from+1, from+2
			if b.Squares[f1].Type == Empty && b.Squares[g1].Type == Empty {
				moves = append(moves, Move{From: from, To: g1})
			}
		}
		if canQueenside {
			d1, c1, b1 := from-1, from-2, from-3
			if b.Squares[d1].Type == Empty && b.Squares[c1].Type == Empty && b.Squares[b1].Type == Empty {
				moves = append(moves, Move{From: from, To: c1})
			}
		}
	}

	return moves
}

func (b *Board) applyMove(m Move) (*Board, bool) {
	nb := *b
	piece := nb.Squares[m.From]

	captured := nb.Squares[m.To]
	gameOver := captured.Type == King

	nb.Squares[m.To] = piece
	nb.Squares[m.From] = Piece{Empty, NoColor}

	if m.Promotion != Empty {
		nb.Squares[m.To] = Piece{m.Promotion, piece.Color}
	}

	if piece.Type == King {
		df := fileOf(m.To) - fileOf(m.From)
		if abs(df) == 2 {
			rankBase := rankOf(m.From) * 8
			if df > 0 {
				nb.Squares[rankBase+5] = nb.Squares[rankBase+7]
				nb.Squares[rankBase+7] = Piece{Empty, NoColor}
			} else {
				nb.Squares[rankBase+3] = nb.Squares[rankBase+0]
				nb.Squares[rankBase+0] = Piece{Empty, NoColor}
			}
		}
		if piece.Color == White {
			nb.WhiteKingside = false
			nb.WhiteQueenside = false
		} else {
			nb.BlackKingside = false
			nb.BlackQueenside = false
		}
	}

	if piece.Type == Rook {
		switch m.From {
		case 56:
			nb.WhiteQueenside = false
		case 63:
			nb.WhiteKingside = false
		case 0:
			nb.BlackQueenside = false
		case 7:
			nb.BlackKingside = false
		}
	}

	nb.EnPassantTarget = "-"

	return &nb, gameOver
}

func (g *GameState) DoMove(m Move) (gameOver bool, err error) {
	p := g.Board.Squares[m.From]

	if p.Type == Empty {
		return false, errors.New("ход из пустой клетки")
	}
	if p.Type == Invisible {
		return false, errors.New("ход из невидимой клетки")
	}
	if p.Color != g.Board.SideToMove {
		return false, errors.New("сейчас ход другого цвета")
	}

	legal := g.Board.pseudoLegalMoves(m.From)
	found := false
	for _, lm := range legal {
		if lm.From == m.From && lm.To == m.To && lm.Promotion == m.Promotion {
			found = true
			break
		}
	}
	if !found {
		return false, errors.New("нелегальный ход")
	}

	nb, over := g.Board.applyMove(m)

	nb.HalfMove++
	if p.Type == Pawn || g.Board.Squares[m.To].Type != Empty {
		nb.HalfMove = 0
	}
	if g.Board.SideToMove == Black {
		nb.FullMove++
	}
	nb.SideToMove = opponent(g.Board.SideToMove)

	g.Board = *nb
	g.Moves = append(g.Moves, m)
	return over, nil
}
