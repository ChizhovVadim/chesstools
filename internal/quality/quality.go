package quality

import (
	"iter"
	"log"

	"github.com/ChizhovVadim/chesstools/pkg/chess"
)

type IProbEvaluator interface {
	// вероятность выигрыша белыми
	EvaluateProb(p *chess.Position) float64
}

func CheckEvalQuality(
	e IProbEvaluator,
	dataset iter.Seq2[DatasetItem, error],
) error {
	var totalCost float64
	var count int

	for item, err := range dataset {
		if err != nil {
			return err
		}
		var x = e.EvaluateProb(&item.Position) - item.Target
		totalCost += x * x
		count += 1
	}

	var averageCost = totalCost / float64(count)
	log.Println("Dataset size:", count)
	log.Printf("Average cost: %f", averageCost)
	return nil
}
