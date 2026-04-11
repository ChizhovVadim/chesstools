package uci

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/ChizhovVadim/chesstools/pkg/chess"
	"github.com/ChizhovVadim/chesstools/pkg/game"
)

var errBadEngineOutput = errors.New("BadEngineOutput")

type Service struct {
	r *bufio.Scanner
	w io.Writer
}

func NewService(r io.Reader, w io.Writer) *Service {
	return &Service{
		r: bufio.NewScanner(r),
		w: w,
	}
}

func (s *Service) Quit() {
	fmt.Fprintln(s.w, "quit")
}

func (s *Service) Uci() EngineInfo {
	fmt.Fprintln(s.w, "uci")
	for s.r.Scan() {
		var msg = s.r.Text()
		if msg == "uciok" {
			return EngineInfo{}
		}
	}
	// TODO err
	return EngineInfo{}
}

func (s *Service) SetOption(option Option) {
	fmt.Fprintf(s.w, "setoption name %v value %v\n",
		option.Name, option.Value)
}

func (s *Service) UciNewgame() {
	fmt.Fprintln(s.w, "ucinewgame")
}

func (s *Service) IsReady() bool {
	fmt.Fprintln(s.w, "isready")
	for s.r.Scan() {
		var msg = s.r.Text()
		if msg == "readyok" {
			return true
		}
	}
	return false
}

func (s *Service) Stop() {
	fmt.Fprintln(s.w, "stop")
}

func (s *Service) Position(game *game.Game) {
	if game.StartFen == "" || game.StartFen == chess.InitialPositionFen {
		fmt.Fprintf(s.w, "position startpos")
	} else {
		fmt.Fprintf(s.w, "position fen %v", game.StartFen)
	}
	if len(game.Moves) > 0 {
		fmt.Fprintf(s.w, " moves")
		for i := range game.Moves {
			var m = game.Moves[i].Move
			fmt.Fprintf(s.w, " %v", m)
		}
	}
	fmt.Fprintln(s.w)
}

func (s *Service) Go(
	game *game.Game,
	tc LimitsType,
	progress func(SearchInfo),
) (SearchResult, error) {
	fmt.Fprintf(s.w, "go ")
	if tc.MoveTime != 0 {
		fmt.Fprintf(s.w, "movetime %v", tc.MoveTime)
	} else if tc.Nodes != 0 {
		fmt.Fprintf(s.w, "nodes %v", tc.Nodes)
	}
	fmt.Fprintln(s.w)

	var lastSearchInfo SearchInfo
	for s.r.Scan() {
		var msg = s.r.Text()
		if msg == "" {
			continue
		}
		var tokens = Tokens{items: strings.Fields(msg)}
		if tokens.Scan() {
			switch tokens.Text() {
			case "info":
				var si = parseSearchInfo(tokens)
				if si.Depth != 0 {
					lastSearchInfo = si
					if progress != nil {
						progress(si)
					}
				}
			case "bestmove":
				if !tokens.Scan() {
					return SearchResult{}, errBadEngineOutput
				}
				var bestMove = game.Position.ParseMoveLAN(tokens.Text())
				if bestMove == chess.MoveEmpty {
					return SearchResult{}, errBadEngineOutput
				}
				return SearchResult{
					BestMove:   bestMove,
					SearchInfo: lastSearchInfo,
				}, nil
			}
		}
	}

	return SearchResult{}, s.r.Err()
}

func parseSearchInfo(tokens Tokens) SearchInfo {
	var res SearchInfo
	for tokens.Scan() {
		var name = tokens.Text()
		if name == "depth" {
			if tokens.Scan() {
				res.Depth, _ = strconv.Atoi(tokens.Text())
			}
		} else if name == "nodes" {
			if tokens.Scan() {
				res.Nodes, _ = strconv.ParseInt(tokens.Text(), 10, 64)
			}
		} else if name == "time" {
			if tokens.Scan() {
				var t, _ = strconv.Atoi(tokens.Text())
				res.Time = time.Duration(t) * time.Millisecond
			}
		} else if name == "score" {
			if tokens.Scan() {
				if tokens.Text() == "cp" {
					if tokens.Scan() {
						var v, _ = strconv.Atoi(tokens.Text())
						res.Score = Score{Centipawns: v}
					}
				} else if tokens.Text() == "mate" {
					if tokens.Scan() {
						var v, _ = strconv.Atoi(tokens.Text())
						res.Score = Score{Mate: v}
					}
				}
			}
		} else if name == "pv" {
			/*for tokens.Scan() {
				common.ParseMoveLAN() tokens.Text()
			}*/
		}
	}
	return res
}
