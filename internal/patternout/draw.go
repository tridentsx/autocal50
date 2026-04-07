package patternout

import (
	"autocal50/internal/pattern"
	"math"
)

// DrawRGBA renders a pattern into a 32-bit XRGB8888 pixel buffer.
// Stride is bytes per row. Pixels are packed as [B, G, R, X] (little-endian XRGB).
func DrawRGBA(buf []byte, w, h, stride int, p pattern.Pattern) {
	// Clear to black.
	for i := range buf {
		buf[i] = 0
	}

	switch p.Type {
	case "solid":
		if p.Color == nil {
			return
		}
		fillRect(buf, w, h, stride, 0, 0, w, h, *p.Color)

	case "window":
		if p.BgColor != nil {
			fillRect(buf, w, h, stride, 0, 0, w, h, *p.BgColor)
		}
		if p.Color == nil {
			return
		}
		pct := p.WindowPc
		if pct <= 0 {
			pct = 18
		}
		side := math.Sqrt(pct / 100)
		ww := int(float64(w) * side)
		wh := int(float64(h) * side)
		x0 := (w - ww) / 2
		y0 := (h - wh) / 2
		fillRect(buf, w, h, stride, x0, y0, ww, wh, *p.Color)

	case "gradient":
		if len(p.Steps) == 0 {
			return
		}
		n := len(p.Steps)
		for x := 0; x < w; x++ {
			idx := x * n / w
			if idx >= n {
				idx = n - 1
			}
			c := p.Steps[idx]
			for y := 0; y < h; y++ {
				off := y*stride + x*4
				buf[off] = c.B
				buf[off+1] = c.G
				buf[off+2] = c.R
				buf[off+3] = 0xFF
			}
		}

	case "grid":
		if p.BgColor != nil {
			fillRect(buf, w, h, stride, 0, 0, w, h, *p.BgColor)
		}
		c := pattern.RGB{R: 255, G: 255, B: 255}
		if p.Color != nil {
			c = *p.Color
		}
		cols := p.Cols
		if cols <= 0 {
			cols = 16
		}
		rows := p.Rows
		if rows <= 0 {
			rows = 9
		}
		for i := 0; i <= cols; i++ {
			x := i * w / cols
			drawVLine(buf, h, stride, x, c)
		}
		for i := 0; i <= rows; i++ {
			y := i * h / rows
			drawHLine(buf, w, stride, y, c)
		}

	case "checker":
		cols := p.Cols
		if cols <= 0 {
			cols = 6
		}
		rows := p.Rows
		if rows <= 0 {
			rows = 4
		}
		cw := w / cols
		ch := h / rows
		for r := 0; r < rows; r++ {
			for col := 0; col < cols; col++ {
				idx := r*cols + col
				c := pattern.RGB{}
				if idx < len(p.Steps) {
					c = p.Steps[idx]
				}
				fillRect(buf, w, h, stride, col*cw, r*ch, cw, ch, c)
			}
		}
	}
}

func fillRect(buf []byte, bw, bh, stride, x, y, rw, rh int, c pattern.RGB) {
	for dy := 0; dy < rh; dy++ {
		py := y + dy
		if py < 0 || py >= bh {
			continue
		}
		rowOff := py * stride
		for dx := 0; dx < rw; dx++ {
			px := x + dx
			if px < 0 || px >= bw {
				continue
			}
			off := rowOff + px*4
			buf[off] = c.B
			buf[off+1] = c.G
			buf[off+2] = c.R
			buf[off+3] = 0xFF
		}
	}
}

func drawVLine(buf []byte, h, stride, x int, c pattern.RGB) {
	for y := 0; y < h; y++ {
		off := y*stride + x*4
		buf[off] = c.B
		buf[off+1] = c.G
		buf[off+2] = c.R
		buf[off+3] = 0xFF
	}
}

func drawHLine(buf []byte, w, stride, y int, c pattern.RGB) {
	off := y * stride
	for x := 0; x < w; x++ {
		buf[off+x*4] = c.B
		buf[off+x*4+1] = c.G
		buf[off+x*4+2] = c.R
		buf[off+x*4+3] = 0xFF
	}
}

// DrawRGBA10 renders a pattern into a 10-bit R10G10B10A2 pixel buffer.
// Each pixel is a uint32: [R:10][G:10][B:10][A:2] packed little-endian.
func DrawRGBA10(buf []byte, w, h, stride int, p pattern.Pattern) {
	// Render 8-bit first into a temp buffer, then upscale.
	tmp := make([]byte, stride*h)
	DrawRGBA(tmp, w, h, stride, p)

	// Convert BGRX8888 → R10G10B10A2.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			off := y*stride + x*4
			b := uint32(tmp[off])
			g := uint32(tmp[off+1])
			r := uint32(tmp[off+2])
			// Scale 8-bit [0..255] to 10-bit [0..1023]: v*1023/255 ≈ v*4+v/64
			r10 := (r*1023 + 127) / 255
			g10 := (g*1023 + 127) / 255
			b10 := (b*1023 + 127) / 255
			// Pack as R10G10B10A2 (little-endian uint32).
			pixel := r10 | (g10 << 10) | (b10 << 20) | (3 << 30)
			buf[off] = byte(pixel)
			buf[off+1] = byte(pixel >> 8)
			buf[off+2] = byte(pixel >> 16)
			buf[off+3] = byte(pixel >> 24)
		}
	}
}
