package main

type UrlMatch struct {
	Url          string
	PercentMatch float32
}

type UrlMatchHeap []UrlMatch

func (h UrlMatchHeap) getTopN(n int) []UrlMatch {
	if n >= h.Len() {
		return h
	} else {
		return h[:n]
	}
}

func (h UrlMatchHeap) Len() int {
	return len(h)
}

// slightly hacky, but the golang heap package expects a "Less" method to make a min heap,
// but we want a max heap (where the highest percent match is the root node)
// so we need to use ">" instead if "<"
func (h UrlMatchHeap) Less(i, j int) bool {
	return h[i].PercentMatch > h[j].PercentMatch
}

func (h UrlMatchHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *UrlMatchHeap) Push(m any) {
	*h = append(*h, m.(UrlMatch))
}

// Pop heap interface requires we have a pop method
// But this case doesn't need a pop method, so it can be a no-op
func (h *UrlMatchHeap) Pop() any { panic("not implemented!") }
