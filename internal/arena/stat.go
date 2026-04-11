package arena

import "math"

type GameStatistics struct {
	winningFraction float64
	eloDifference   float64
	los             float64
}

// https://chessprogramming.wikispaces.com/Match%20Statistics
func computeStat(wins, losses, draws int) GameStatistics {
	var games = wins + losses + draws
	var winning_fraction = (float64(wins) + 0.5*float64(draws)) / float64(games)
	var elo_difference = -math.Log(1/winning_fraction-1) * 400 / math.Ln10
	var los = 0.5 + 0.5*math.Erf(float64(wins-losses)/math.Sqrt(2*float64(wins+losses)))
	return GameStatistics{
		winningFraction: winning_fraction,
		eloDifference:   elo_difference,
		los:             los,
	}
}
