package lexer

import "io"

type runeQueue struct {
	queue []rune
	size  uint8
	head  uint8 // ptr to the next head elem
	tail  uint8 // ptr to the current tail elem
}

func newRuneQueue(size uint8) *runeQueue {
	return &runeQueue{
		queue: make([]rune, size+1, size+1),
		size:  size,
		head:  0,
		tail:  0,
	}
}

func (rq *runeQueue) Size() uint8 {
	size := (rq.tail - rq.head) % (rq.size + 1)
	if size < 0 {
		return rq.size + size
	}
	return size
}

func (rq *runeQueue) Head(n uint8) rune {
	if rq.Size() < n {
		panic("shouldn't happen!")
	}
	return rq.queue[(rq.head+n)%(rq.size+1)]
}

func (rq *runeQueue) Enqueue(x rune) {
	if rq.Size() == rq.size {
		panic("shouldn't happen!")
	}
	rq.queue[rq.tail] = x
	rq.tail = (rq.tail + 1) % (rq.size + 1)
}

func (rq *runeQueue) Dequeue() rune {
	x := rq.Head(0)
	rq.head = (rq.head + 1) % (rq.size + 1)
	return x
}

// peeker is a helper to maintain a lookahead
// into the next runes, and keep track of our
// position in the stream.
type peeker struct {
	r      io.RuneReader
	rq     runeQueue
	prev   rune // from the previous call to Next()
	lineNo int
	column int
	sz     uint8 // when p.err != nil && p.rq.Size() == sz, return p.err
	err    error // any error during reading from r
}

func newPeeker(r io.RuneReader, size uint8) *peeker {
	p := &peeker{}
	p.r = r
	p.rq = *newRuneQueue(size)
	p.lineNo = 1
	p.err = nil
	return p
}

func (p *peeker) Next() (rune, error) {
	r, err := p.Peek(0)
	if err == nil {
		p.rq.Dequeue()
	}
	switch p.prev {
	case '\n':
		p.lineNo++
		p.column = 1
	default:
		p.column++
	}
	p.prev = r
	return r, err
}

func (p *peeker) Peek(count uint8) (rune, error) {
	sz := p.rq.Size()
	if p.err != nil && sz == p.sz {
		return 0, p.err
	}
	for count >= sz {
		r, _, err := p.r.ReadRune()
		if err != nil {
			p.sz = sz
			p.err = err
			return r, err
		}
		p.rq.Enqueue(r)
		sz++
	}
	// we only enqueue on a successful read, so
	// this must have been from a successful read.
	return p.rq.Head(count), nil
}
