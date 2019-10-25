package ledgerstate

type Color [colorLength]byte

func NewColor(color string) (result Color) {
	copy(result[:], color)

	return
}

func (color Color) String() string {
	return "color"
}

const colorLength = 8
