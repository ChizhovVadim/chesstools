package pgn

import (
	"bufio"
	"iter"
	"os"
	"strings"
)

type GameRaw struct {
	Tags    []Tag
	BodyRaw string
}

type Tag struct {
	Key   string
	Value string
}

// https://en.wikipedia.org/wiki/Portable_Game_Notation
func LoadGames(path string) iter.Seq2[GameRaw, error] {
	return func(yield func(GameRaw, error) bool) {
		file, err := os.Open(path)
		if err != nil {
			yield(GameRaw{}, err)
			return
		}
		defer file.Close()

		var tags []string
		var body = &strings.Builder{}
		var hasBody bool

		var scanner = bufio.NewScanner(file)
		for scanner.Scan() {
			var line = scanner.Text()
			if strings.HasPrefix(line, "[") {
				if hasBody {
					if len(tags) != 0 && body.Len() != 0 {
						if !yield(GameRaw{
							Tags:    parseTags(tags),
							BodyRaw: body.String(),
						}, nil) {
							return
						}
					}
					hasBody = false
					tags = nil
					body.Reset()
				}
				tags = append(tags, line)
			} else {
				hasBody = true
				body.WriteString(line)
				body.WriteString(" ")
			}
		}
		if hasBody && len(tags) != 0 && body.Len() != 0 {
			if !yield(GameRaw{
				Tags:    parseTags(tags),
				BodyRaw: body.String(),
			}, nil) {
				return
			}
		}
	}
}

func (g *GameRaw) TagValue(key string) string {
	for _, tag := range g.Tags {
		if tag.Key == key {
			return tag.Value
		}
	}
	return ""
}

func parseTags(tags []string) []Tag {
	var result []Tag
	for _, tag := range tags {
		tag = strings.TrimLeft(tag, "[")
		tag = strings.TrimRight(tag, "]")
		var i0 = strings.Index(tag, "\"")
		var i1 = strings.LastIndex(tag, "\"")
		if i0 == -1 || i1 == -1 {
			continue
		}
		var name = strings.TrimSpace(tag[:i0])
		var val = tag[i0+1 : i1]
		result = append(result, Tag{Key: name, Value: val})
	}
	return result
}
