package quality

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/ChizhovVadim/chesstools/pkg/chess"
)

type Service struct {
	r            *bufio.Scanner
	w            io.Writer
	sigmoidScale float64
}

func NewService(
	r io.Reader,
	w io.Writer,
	sigmoidScale float64,
) *Service {
	return &Service{
		r:            bufio.NewScanner(r),
		w:            w,
		sigmoidScale: sigmoidScale,
	}
}

func (s *Service) Evaluate(p *chess.Position) int {
	fmt.Fprintln(s.w, "position fen", p.String())
	fmt.Fprintln(s.w, "eval")
	if s.r.Scan() {
		var msg = s.r.Text()
		var f, err = strconv.Atoi(msg)
		if err == nil {
			return f
		}
	}
	return 0
}

func (s *Service) EvaluateProb(p *chess.Position) float64 {
	var staticEval = s.Evaluate(p)
	return sigmoid(s.sigmoidScale * float64(staticEval))
}

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}
