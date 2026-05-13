package chessengine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Forsyth–Edwards Notation (FEN)
// Пример: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
//
// Поля FEN (разделены пробелами):
// 1. Позиция фигур — ряды с 8-го по 1-й, разделённые «/».
//    Цифра означает количество пустых клеток подряд.
//    Заглавные буквы — белые фигуры, строчные — чёрные.
// 2. Очерёдность хода: w — ход белых, b — ход чёрных.
// 3. Доступность рокировки:
//    K — белые могут рокировать в короткую сторону (королевский фланг)
//    Q — белые могут рокировать в длинную сторону (ферзевый фланг)
//    k/q — то же самое для чёрных
//    «-» — рокировка недоступна ни для кого
// 4. Взятие на проходе: поле, на которое можно взять пешку на проходе,
//    например «e6». «-» — взятие на проходе невозможно.
// 5. Счётчик полуходов — число полуходов с последнего взятия или хода пешки.
//    При достижении 50 фиксируется ничья по правилу 50 ходов.
// 6. Номер полного хода — увеличивается после каждого хода чёрных.
//
// Обозначения фигур:
// ♔ K = Король  (белый)
// ♕ Q = Ферзь   (белый)
// ♖ R = Ладья   (белый)
// ♗ B = Слон    (белый)
// ♘ N = Конь    (белый)
// ♙ P = Пешка   (белая)
// ♚ k = Король  (чёрный)
// ♛ q = Ферзь   (чёрный)
// ♜ r = Ладья   (чёрная)
// ♝ b = Слон    (чёрный)
// ♞ n = Конь    (чёрный)
// ♟ p = Пешка   (чёрная)
// *   = невидимое поле

type Color uint8

const (
	White Color = iota
	Black
	NoColor
)

type PieceType uint8

const (
	Empty PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
	Invisible
)

type Board struct {
	Squares         [64]Piece
	SideToMove      Color
	WhiteKingside   bool
	WhiteQueenside  bool
	BlackKingside   bool
	BlackQueenside  bool
	EnPassantTarget string
	HalfMove        int
	FullMove        int
}

type Piece struct {
	Type  PieceType
	Color Color
}

func pieceToChar(p Piece) rune {
	if p.Type == Empty {
		return 0
	}
	ch := ""
	switch p.Type {
	case Invisible:
		ch = "*"
	case Pawn:
		ch = "p"
	case Knight:
		ch = "n"
	case Bishop:
		ch = "b"
	case Rook:
		ch = "r"
	case Queen:
		ch = "q"
	case King:
		ch = "k"
	}
	if p.Color == White {
		ch = strings.ToUpper(ch)
	}
	return rune(ch[0])
}

var figureMap = map[rune]PieceType{'*': Invisible,
	'P': Pawn, 'N': Knight, 'B': Bishop, 'R': Rook, 'Q': Queen, 'K': King,
	'p': Pawn, 'n': Knight, 'b': Bishop, 'r': Rook, 'q': Queen, 'k': King,
}

func parseFen(fen string) (*Board, error) {
	parts := strings.Split(fen, " ")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid FEN: expected 6 parts, got %d", len(parts))
	}

	board := &Board{}
	rows := strings.Split(parts[0], "/")
	if len(rows) != 8 {
		return nil, fmt.Errorf("invalid board part: expected 8 rows, got %d", len(rows))
	}

	pointer := 0

	for _, row := range rows {
		for _, ch := range row {
			if ch >= '1' && ch <= '8' {
				empty := int(ch - '0')
				for i := 0; i < empty; i++ {
					board.Squares[pointer] = Piece{Empty, NoColor}
					pointer++
				}
			} else if ch == '*' {
				board.Squares[pointer] = Piece{Invisible, NoColor}
				pointer++
			} else {
				pieceType, ok := figureMap[ch]
				if !ok {
					return nil, fmt.Errorf("invalid piece symbol: %c", ch)
				}
				color := White
				if ch >= 'a' && ch <= 'z' {
					color = Black
				}
				board.Squares[pointer] = Piece{pieceType, color}
				pointer++
			}
		}
		if pointer > 64 {
			return nil, errors.New("too many squares in FEN")
		}
	}
	if pointer != 64 {
		return nil, fmt.Errorf("invalid board: got %d squares, expected 64", pointer)
	}

	switch parts[1] {
	case "w":
		board.SideToMove = White
	case "b":
		board.SideToMove = Black
	default:
		return nil, fmt.Errorf("invalid side to move: %s", parts[1])
	}

	castle := parts[2]
	if castle == "-" {
		board.WhiteKingside, board.WhiteQueenside = false, false
		board.BlackKingside, board.BlackQueenside = false, false
	} else {
		for _, ch := range castle {
			switch ch {
			case 'K':
				board.WhiteKingside = true
			case 'Q':
				board.WhiteQueenside = true
			case 'k':
				board.BlackKingside = true
			case 'q':
				board.BlackQueenside = true
			default:
				return nil, fmt.Errorf("invalid castle symbol: %c", ch)
			}
		}
	}

	board.EnPassantTarget = parts[3]
	if board.EnPassantTarget != "-" && len(board.EnPassantTarget) != 2 {
		return nil, fmt.Errorf("invalid en passant target: %s", board.EnPassantTarget)
	}

	halfMove, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, fmt.Errorf("invalid halfmove: %s", parts[4])
	}
	board.HalfMove = halfMove

	fullMove, err := strconv.Atoi(parts[5])
	if err != nil {
		return nil, fmt.Errorf("invalid fullmove: %s", parts[5])
	}
	board.FullMove = fullMove

	return board, nil
}

func (b *Board) formatFen() (string, error) {
	var sb strings.Builder
	sb.Grow(128)

	for row := 0; row < 8; row++ {
		empty := 0
		for col := 0; col < 8; col++ {
			idx := row*8 + col
			piece := b.Squares[idx]
			if piece.Type == Empty {
				empty++
			} else {
				if empty > 0 {
					sb.WriteByte(byte('0' + empty))
					empty = 0
				}
				ch := pieceToChar(piece)
				sb.WriteRune(ch)
			}
		}
		if empty > 0 {
			sb.WriteByte(byte('0' + empty))
		}
		if row != 7 {
			sb.WriteByte('/')
		}
	}

	sb.WriteByte(' ')
	if b.SideToMove == White {
		sb.WriteByte('w')
	} else {
		sb.WriteByte('b')
	}

	sb.WriteByte(' ')
	castle := ""
	if b.WhiteKingside {
		castle += "K"
	}
	if b.WhiteQueenside {
		castle += "Q"
	}
	if b.BlackKingside {
		castle += "k"
	}
	if b.BlackQueenside {
		castle += "q"
	}
	if castle == "" {
		castle = "-"
	}
	sb.WriteString(castle)

	sb.WriteByte(' ')
	if b.EnPassantTarget == "" {
		sb.WriteByte('-')
	} else {
		sb.WriteString(b.EnPassantTarget)
	}

	sb.WriteString(fmt.Sprintf(" %d %d", b.HalfMove, b.FullMove))

	return sb.String(), nil
}

var pieceToSymbol = map[Piece]string{
	{Empty, NoColor}:     "·",
	{Invisible, NoColor}: "*",

	{Pawn, White}:   "♙",
	{Knight, White}: "♘",
	{Bishop, White}: "♗",
	{Rook, White}:   "♖",
	{Queen, White}:  "♕",
	{King, White}:   "♔",

	{Pawn, Black}:   "♟",
	{Knight, Black}: "♞",
	{Bishop, Black}: "♝",
	{Rook, Black}:   "♜",
	{Queen, Black}:  "♛",
	{King, Black}:   "♚",
}

func (b *Board) Print() {
	fmt.Println("  a b c d e f g h")
	for row := 0; row < 8; row++ {
		fmt.Printf("%d ", 8-row)
		for col := 0; col < 8; col++ {
			idx := row*8 + col
			piece := b.Squares[idx]
			sym, ok := pieceToSymbol[piece]
			if !ok {
				sym = "?"
			}
			fmt.Printf("%s ", sym)
		}
		fmt.Printf("%d\n", 8-row)
	}
	fmt.Println("  a b c d e f g h")
}

var letterToInt = map[rune]int{
	'a': 0,
	'b': 1,
	'c': 2,
	'd': 3,
	'e': 4,
	'f': 5,
	'g': 6,
	'h': 7,
}

var intToLetter = "abcdefgh"

func ConverToIndex(s string) (int, error) {
	if len(s) != 2 {
		return 0, errors.New("invalid square")
	}
	col, ok := letterToInt[rune(s[0])]
	if !ok {
		return 0, errors.New("invalid file")
	}
	row, err := strconv.Atoi(string(s[1]))
	if err != nil || row < 1 || row > 8 {
		return 0, errors.New("invalid rank")
	}
	// индекс: 0 = a8, 63 = h1
	return (8-row)*8 + col, nil
}

func IndexToCoord(idx int) string {
	row := 8 - idx/8
	col := idx % 8
	return string(intToLetter[col]) + strconv.Itoa(row)
}

func (b *Board) FenForPlayer(color Color) (string, error) {
	fogBoard := b.BoardForPlayer(color)
	if color == White {
		fogBoard.BlackKingside = false
		fogBoard.BlackQueenside = false
	} else {
		fogBoard.WhiteKingside = false
		fogBoard.WhiteQueenside = false
	}
	return fogBoard.formatFen()
}
