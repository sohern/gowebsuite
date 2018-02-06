package structures

// Page structure for wiki
type Page struct {
	Title string `json:"title,omitempty"`
	Body  []byte `json:"body,omitempty"`
}
