package image

// Model for a polymorphic image, being used by different other models.
type Image struct {
	ID     uint   `json:"id"`
	Full   string `json:"fullImage"`
	Sprite string `json:"sprite"`
	X      uint16 `json:"x"`
	Y      uint16 `json:"y"`
	W      uint16 `json:"w"`
	H      uint16 `json:"h"`
}
