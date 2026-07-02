package tray

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
)

// makeIcon generates a 22x22 template PNG — a sine wave in black on transparent.
// macOS tints template images automatically to match the menu bar colour scheme.
func makeIcon() []byte {
	const size = 22
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	black := color.NRGBA{R: 0, G: 0, B: 0, A: 255}

	// Draw two sine wave cycles across the icon, vertically centred
	for x := 0; x < size; x++ {
		t := float64(x) / float64(size) * 2 * math.Pi * 1.5
		y := int(math.Round(float64(size)/2 - math.Sin(t)*4.5))
		for dy := -1; dy <= 1; dy++ {
			py := y + dy
			if py >= 0 && py < size {
				alpha := uint8(255)
				if dy != 0 {
					alpha = 120
				}
				img.SetNRGBA(x, py, color.NRGBA{R: 0, G: 0, B: 0, A: alpha})
			}
		}
	}
	_ = black // used via SetNRGBA above

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
