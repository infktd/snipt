package components

// SetShowPreview enables or disables the two-column preview layout.
func (r *ResultList) SetShowPreview(v bool) {
	r.showPreview = v
}

// SetWidth sets the display width of the result list.
func (r *ResultList) SetWidth(w int) {
	r.width = w
}

// SetHeight sets the maximum number of visible rows.
func (r *ResultList) SetHeight(h int) {
	r.height = h
}
