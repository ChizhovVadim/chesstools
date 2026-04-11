package uci

type Tokens struct {
	items   []string
	index   int
	current string
}

func (t *Tokens) Scan() bool {
	if t.index < len(t.items) {
		t.current = t.items[t.index]
		t.index += 1
		return true
	} else {
		t.current = ""
		return false
	}
}

func (t *Tokens) Text() string {
	return t.current
}
