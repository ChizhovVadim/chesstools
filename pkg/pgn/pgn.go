package pgn

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/ChizhovVadim/chesstools/pkg/chess"
	"github.com/ChizhovVadim/chesstools/pkg/game"
)

type Comment struct {
	Score     game.Score
	Depth     int
	IsOpening bool
}

func GameResultName(gameResult int) string {
	switch gameResult {
	case game.GameResultDraw:
		return "1/2-1/2"
	case game.GameResultWhiteWins:
		return "1-0"
	case game.GameResultBlackWins:
		return "0-1"
	default:
		return "*"
	}
}

func ParseGameResult(s string) int {
	switch s {
	case "1/2-1/2":
		return game.GameResultDraw
	case "1-0":
		return game.GameResultWhiteWins
	case "0-1":
		return game.GameResultBlackWins
	default:
		return game.GameResultNone
	}
}

func ParseGame(gameRaw GameRaw) (game.Game, error) {
	var result, err = game.NewGame(gameRaw.TagValue("FEN"))
	if err != nil {
		return game.Game{}, err
	}
	var tokens = parsePgnBody(gameRaw.BodyRaw)
	for i := range tokens {
		var san = tokens[i].Value
		var move = chess.ParseMoveSAN(&result.Position, san)
		if move == chess.MoveEmpty {
			break
		}
		var comment = parseComment(tokens[i].Comment)
		result.MakeMove(game.MoveItem{
			Move:      move,
			Score:     comment.Score,
			Depth:     comment.Depth,
			IsOpening: comment.IsOpening,
		})
	}
	result.Result = ParseGameResult(gameRaw.TagValue("Result"))
	return result, nil
}

// TODO корректно обрабатывать здесь ошибки - боль
func Write(g *game.Game, w io.Writer) error {
	// order: Date, Round, White, Black, Result, FEN
	if !g.Date.IsZero() {
		writePgnTag(w, "Date", g.Date.Format("2006.01.02"))
	}
	if g.Round != 0 {
		writePgnTag(w, "Round", strconv.Itoa(g.Round))
	}
	writePgnTag(w, "White", g.White)
	writePgnTag(w, "Black", g.Black)
	writePgnTag(w, "Result", GameResultName(g.Result))
	if !(g.StartFen == "" || g.StartFen == chess.InitialPositionFen) {
		writePgnTag(w, "FEN", g.StartFen)
	}
	fmt.Fprintln(w)
	if err := WriteMoves(w, g, true); err != nil {
		return err
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w)
	return nil
}

func WriteMoves(w io.Writer, g *game.Game, comments bool) error {
	curPos, err := chess.NewPositionFromFEN(g.StartFen)
	if err != nil {
		return err
	}
	for i, item := range g.Moves {
		var child chess.Position
		curPos.MakeMove(item.Move, &child)

		// TODO номера ходов, шахи, маты, результат игры в конце.
		// TODO в одну строку или новую строку после N ходов?
		if i%2 == 0 {
			fmt.Fprintf(w, "%v. ", i/2+1)
		}
		var san = chess.MoveToSan(&curPos, item.Move)
		if child.IsCheck() {
			if i == len(g.Moves)-1 && g.ResultComment == "checkmate" {
				san += "#"
			} else {
				san += "+"
			}
		}
		fmt.Fprintf(w, "%v ", san)
		if comments {
			writeComment(w, item)
		}

		curPos = child
	}
	fmt.Fprintf(w, "%v\n", GameResultName(g.Result))
	return nil
}

func writeComment(w io.Writer, item game.MoveItem) {
	if item.IsOpening {
		fmt.Fprintf(w, "{book} ")
		return
	}
	if item.Depth == 0 {
		return
	}
	// С точки зрения какой стороны пишем оценку?
	if item.Score.Mate != 0 {
		var signChar rune
		var mate = item.Score.Mate
		if mate > 0 {
			signChar = '+'
		} else {
			signChar = '-'
			mate = -mate
		}
		fmt.Fprintf(w, "{%cM%v/%v} ", signChar, mate, item.Depth)
	} else {
		var score = 0.01 * float64(item.Score.Centipawns)
		fmt.Fprintf(w, "{%+.2f/%v} ", score, item.Depth)
	}
}

func writePgnTag(w io.Writer, key, value string) {
	fmt.Fprintf(w, "[%v \"%v\"]\n", key, value)
}

type token struct {
	Value   string
	Comment string
}

func parsePgnBody(bodyRaw string) []token {
	var result []token
	var inComment = false
	var body string
	for _, rune := range bodyRaw {
		if inComment {
			if rune == '}' {
				if len(result) != 0 {
					result[len(result)-1].Comment = body
				}
				inComment = false
				body = ""
			} else {
				body = body + string(rune)
			}
		} else if rune == '.' {
			body = ""
		} else if unicode.IsSpace(rune) {
			if body != "" {
				result = append(result, token{Value: body})
				body = ""
			}
		} else if rune == '{' {
			if body != "" {
				result = append(result, token{Value: body})
				body = ""
			}
			inComment = true
			body = ""
		} else {
			body = body + string(rune)
		}
	}
	if body != "" {
		result = append(result, token{Value: body})
	}
	return result
}

func parseComment(comment string) Comment {
	comment = strings.TrimLeft(comment, "{")
	comment = strings.TrimRight(comment, "}")
	if comment == "book" {
		return Comment{IsOpening: true}
	}
	var fields = strings.Fields(comment)
	if len(fields) == 0 {
		return Comment{}
	}
	var s string
	if len(fields) > 1 && strings.HasPrefix(fields[0], "(") {
		s = fields[1]
	} else {
		s = fields[0]
	}
	if s == "" {
		return Comment{}
	}
	var index = strings.Index(s, "/")
	if index < 0 {
		return Comment{}
	}
	var sScore = s[:index]
	var sDepth = s[index+1:]

	var uciScore game.Score
	if strings.Contains(sScore, "M") {
		sScore = strings.Replace(sScore, "M", "", 1)
		score, err := strconv.Atoi(sScore)
		if err != nil {
			return Comment{}
		}
		uciScore = game.Score{Mate: score}
	} else {
		score, err := strconv.ParseFloat(sScore, 64)
		if err != nil {
			return Comment{}
		}
		uciScore = game.Score{Centipawns: int(100 * score)}
	}

	depth, err := strconv.Atoi(sDepth)
	if err != nil {
		return Comment{}
	}

	return Comment{
		Score: uciScore,
		Depth: depth,
	}
}
