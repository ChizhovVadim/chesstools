package quality

import (
	"bufio"
	"fmt"
	"iter"
	"os"
	"strings"

	"github.com/ChizhovVadim/chesstools/pkg/chess"
)

type DatasetItem struct {
	Position chess.Position
	Target   float64
}

func LoadValidationDataset(path string) iter.Seq2[DatasetItem, error] {
	return func(yield func(DatasetItem, error) bool) {
		file, err := os.Open(path)
		if err != nil {
			if !yield(DatasetItem{}, err) {
				return
			}
		}
		defer file.Close()

		var scanner = bufio.NewScanner(file)
		for scanner.Scan() {
			var s = scanner.Text()
			var item, err = parseItem(s)
			if !yield(item, err) {
				break
			}
		}
	}
}

func parseItem(s string) (DatasetItem, error) {
	var index = strings.Index(s, "\"")
	if index < 0 {
		return DatasetItem{}, fmt.Errorf("zurichessParser failed %v", s)
	}

	var fen = s[:index]
	var pos, err = chess.NewPositionFromFEN(fen)
	if err != nil {
		return DatasetItem{}, err
	}

	var strScore = s[index+1:]

	var prob float64
	if strings.HasPrefix(strScore, "1/2-1/2") {
		prob = 0.5
	} else if strings.HasPrefix(strScore, "1-0") {
		prob = 1.0
	} else if strings.HasPrefix(strScore, "0-1") {
		prob = 0.0
	} else {
		return DatasetItem{}, fmt.Errorf("zurichessParser failed %v", s)
	}
	return DatasetItem{Position: pos, Target: prob}, nil
}
