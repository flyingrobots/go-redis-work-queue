package tui

// layout helpers for view composition

// bodyDims computes the body width and height available for panels,
// accounting for header/sub lines, optional filter line, and footer/status bar.
func (m model) bodyDims() (w, h int) {
	w = m.width
	headerLines := 4 // tab bar + header + sub + blank
	if m.filterActive || trim(m.filter.Value()) != "" {
		headerLines++
	}
	footerLines := 2 // status bar and spacing
	h = m.height - headerLines - footerLines
	if h < 6 {
		h = 6
	}
	return w, h
}

func trim(s string) string {
	// local small helper to avoid importing strings here
	// trims ASCII spaces only, enough for our use
	b := []rune(s)
	i := 0
	j := len(b) - 1
	for i <= j && (b[i] == ' ' || b[i] == '\t' || b[i] == '\n') {
		i++
	}
	for j >= i && (b[j] == ' ' || b[j] == '\t' || b[j] == '\n') {
		j--
	}
	if i > j {
		return ""
	}
	return string(b[i : j+1])
}
