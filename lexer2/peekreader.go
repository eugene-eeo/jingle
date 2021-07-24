package lexer2

import "io"

type runeInfo struct {
	r rune
	i int
	e error
}

func newPeekRuneReader(r io.RuneReader) *peekRuneReader {
	p := &peekRuneReader{src: r}
	_, _, _ = p.ReadRune() // so that the next call to ReadRune is valid...
	return p
}

type peekRuneReader struct {
	src  io.RuneReader
	curr runeInfo
	next runeInfo
}

func (p *peekRuneReader) ReadRune() (rune, int, error) {
	// advance!
	r, i, e := p.src.ReadRune()
	p.curr = p.next
	p.next = runeInfo{r, i, e}
	return p.curr.r, p.curr.i, p.curr.e
}

func (p *peekRuneReader) Peek() (rune, error) {
	return p.next.r, p.next.e
}
