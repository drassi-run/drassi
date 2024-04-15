package parser

const (
	MinObjectSize      = 24
	StringBaseOverhead = 26
)

type MemoryCounter struct {
	node         *ExpressionNode
	maxBytes     int
	currentBytes int
}

func NewMemoryCounter(node *ExpressionNode, maxBytes int) *MemoryCounter {
	return &MemoryCounter{
		node:     node,
		maxBytes: maxBytes,
	}
}
func (m *MemoryCounter) GetCurrentBytes() int {
	return m.currentBytes
}

func (m *MemoryCounter) AddInt(amount int) {

}

func (m *MemoryCounter) AddStr(value string) {

}

func (m *MemoryCounter) AddMinObjSize() {
	m.AddInt(MinObjectSize)
}

func (m *MemoryCounter) Remove(value string) {
	m.currentBytes -= m.CalculateSize(value)
}

func (m *MemoryCounter) CalculateSize(value string) int {
	// This measurement doesn't have to be perfect.
	// https://codeblog.jonskeet.uk/2011/04/05/of-memory-and-strings/
	var b int
	if len(value) == 0 {
		b = StringBaseOverhead
	} else {
		b = StringBaseOverhead + len(value)*2
	}
	return b
}

func (m *MemoryCounter) tryAddInt(amount int) (ok bool) {
	defer func() {
		if err := recover(); err != nil {
			ok = false
		}
	}()
	amount += m.currentBytes
	if amount > m.maxBytes {
		return false
	}
	m.currentBytes = amount
	return true
}

func (m *MemoryCounter) tryAddString(amount string) (ok bool) {
	return m.tryAddInt(m.CalculateSize(amount))
}
