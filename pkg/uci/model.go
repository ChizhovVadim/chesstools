package uci

import (
	"time"

	"github.com/ChizhovVadim/chesstools/pkg/chess"
)

type EngineInfo struct {
	Name    string
	Author  string
	Options []OptionInfo
}

type OptionInfo struct {
	Name string
}

type Option struct {
	Name  string
	Value string
}

type LimitsType struct {
	Ponder         bool
	Infinite       bool
	WhiteTime      int
	BlackTime      int
	WhiteIncrement int
	BlackIncrement int
	MoveTime       int
	MovesToGo      int
	Depth          int
	Nodes          int
	Mate           int
}

type SearchInfo struct {
	Depth    int
	Nodes    int64
	Time     time.Duration
	Score    Score
	MainLine []chess.Move
}

type SearchResult struct {
	SearchInfo
	BestMove chess.Move
}

type Score struct {
	Centipawns int
	Mate       int
}
