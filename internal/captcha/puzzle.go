package captcha

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"math/rand/v2"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // 远端 API（如 alcy）常返回 webp

	_ "image/gif" // register GIF decoder for remote captcha images
)

// PuzzleResult 对齐 mypage create_puzzle_from_bytes 输出。
type PuzzleResult struct {
	BGDataURL     string
	PuzzleDataURL string
	X             int
	Y             int
	BGWidth       int
	BGHeight      int
	PuzzleSize    int
}

// CreatePuzzleFromBytes 从底图字节生成背景缺口图与拼图块（对齐 mypage）。
func CreatePuzzleFromBytes(imageBytes []byte, bgWidth, bgHeight, puzzleSize int) (*PuzzleResult, error) {
	// 1. 尺寸缺省与解码底图，cover-crop 到目标画布
	if bgWidth <= 0 {
		bgWidth = 320
	}
	if bgHeight <= 0 {
		bgHeight = 160
	}
	if puzzleSize <= 0 {
		puzzleSize = 60
	}

	raw, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, fmt.Errorf("decode captcha image: %w", err)
	}
	bg := coverCropRGBA(raw, bgWidth, bgHeight)

	// 2. 计算可放置缺口的随机坐标（留边距给凸耳）
	tab := puzzleSize / 6
	if tab < 8 {
		tab = 8
	}
	margin := tab + 4
	maxX := bgWidth - puzzleSize - margin
	maxY := bgHeight - puzzleSize - margin
	if maxX <= margin || maxY <= margin {
		return nil, fmt.Errorf("captcha canvas too small for puzzle")
	}
	minX := margin + puzzleSize/2
	x := rand.IntN(maxX-minX+1) + minX
	y := rand.IntN(maxY-margin+1) + margin

	// 3. 生成拼图蒙版，裁出滑块并在背景挖半透明缺口
	mask := jigsawMask(puzzleSize, tab)

	piece := image.NewRGBA(image.Rect(0, 0, puzzleSize, puzzleSize))
	draw.Draw(piece, piece.Bounds(), bg, image.Pt(x, y), draw.Src)
	applyAlphaMask(piece, mask)

	darkOverlay(bg, x, y, puzzleSize, mask)
	outlinePiece(piece, mask)

	// 4. 编码为 data URL 返回（背景 JPEG、滑块 PNG 保透明）
	bgURL, err := jpegDataURL(bg, 85)
	if err != nil {
		return nil, err
	}
	pieceURL, err := pngDataURL(piece)
	if err != nil {
		return nil, err
	}
	return &PuzzleResult{
		BGDataURL:     bgURL,
		PuzzleDataURL: pieceURL,
		X:             x,
		Y:             y,
		BGWidth:       bgWidth,
		BGHeight:      bgHeight,
		PuzzleSize:    puzzleSize,
	}, nil
}

func coverCropRGBA(src image.Image, w, h int) *image.RGBA {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	scale := float64(w) / float64(sw)
	if float64(h)/float64(sh) > scale {
		scale = float64(h) / float64(sh)
	}
	rw := max(1, int(float64(sw)*scale))
	rh := max(1, int(float64(sh)*scale))
	resized := image.NewRGBA(image.Rect(0, 0, rw, rh))
	xdraw.CatmullRom.Scale(resized, resized.Bounds(), src, sb, xdraw.Over, nil)

	left := (rw - w) / 2
	top := (rh - h) / 2
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(out, out.Bounds(), resized, image.Pt(left, top), draw.Src)
	return out
}

func jigsawMask(size, tab int) *image.Alpha {
	mask := image.NewAlpha(image.Rect(0, 0, size, size))
	body := size - tab*2
	// 主体方块
	fillRectAlpha(mask, tab, tab, tab+body, tab+body, 255)
	// 右侧凸起
	cx := tab + body
	cy := tab + body/2
	fillEllipseAlpha(mask, cx, cy, tab, tab, 255)
	// 顶部凹槽（挖空）
	tx := tab + body/2
	ty := tab
	fillEllipseAlpha(mask, tx, ty, tab, tab, 0)
	return mask
}

func fillRectAlpha(img *image.Alpha, x0, y0, x1, y1 int, a uint8) {
	b := img.Bounds()
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
				continue
			}
			img.SetAlpha(x, y, color.Alpha{A: a})
		}
	}
}

func fillEllipseAlpha(img *image.Alpha, cx, cy, rx, ry int, a uint8) {
	b := img.Bounds()
	for y := cy - ry; y <= cy+ry; y++ {
		for x := cx - rx; x <= cx+rx; x++ {
			if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
				continue
			}
			dx := float64(x-cx) / float64(rx)
			dy := float64(y-cy) / float64(ry)
			if dx*dx+dy*dy <= 1 {
				img.SetAlpha(x, y, color.Alpha{A: a})
			}
		}
	}
}

func applyAlphaMask(img *image.RGBA, mask *image.Alpha) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			a := mask.AlphaAt(x, y).A
			off := img.PixOffset(x, y)
			img.Pix[off+3] = a
			if a == 0 {
				img.Pix[off+0] = 0
				img.Pix[off+1] = 0
				img.Pix[off+2] = 0
			}
		}
	}
}

func darkOverlay(bg *image.RGBA, x, y, size int, mask *image.Alpha) {
	for dy := 0; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			if mask.AlphaAt(dx, dy).A == 0 {
				continue
			}
			px, py := x+dx, y+dy
			r, g, b, a := bg.At(px, py).RGBA()
			// 叠一层约 110/255 的黑色
			nr := uint8(float64(r>>8) * (1 - 110.0/255))
			ng := uint8(float64(g>>8) * (1 - 110.0/255))
			nb := uint8(float64(b>>8) * (1 - 110.0/255))
			bg.Set(px, py, color.RGBA{R: nr, G: ng, B: nb, A: uint8(a >> 8)})
		}
	}
}

func outlinePiece(piece *image.RGBA, mask *image.Alpha) {
	b := piece.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if mask.AlphaAt(x, y).A == 0 {
				continue
			}
			// 边缘像素叠半透明白
			edge := false
			for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || ny < 0 || nx >= b.Dx() || ny >= b.Dy() || mask.AlphaAt(nx, ny).A == 0 {
					edge = true
					break
				}
			}
			if !edge {
				continue
			}
			off := piece.PixOffset(x, y)
			// alpha composite white@160 over pixel
			srcA := 160.0 / 255
			dr := float64(piece.Pix[off+0])
			dg := float64(piece.Pix[off+1])
			db := float64(piece.Pix[off+2])
			piece.Pix[off+0] = uint8(255*srcA + dr*(1-srcA))
			piece.Pix[off+1] = uint8(255*srcA + dg*(1-srcA))
			piece.Pix[off+2] = uint8(255*srcA + db*(1-srcA))
		}
	}
}

func jpegDataURL(img image.Image, quality int) (string, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return "", err
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func pngDataURL(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// SolidFallbackImage 远端不可用时的兜底底图（对齐 mypage solid_fallback_image）。
func SolidFallbackImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 70, G: 130, B: 180, A: 255}}, image.Point{}, draw.Src)
	for i := 0; i < width; i += 20 {
		shade := uint8(60 + (i*7)%80)
		c := color.RGBA{R: shade, G: shade + 40, B: 200, A: 255}
		for t := 0; t < height; t++ {
			x := i + t
			y := t
			if x >= 0 && x < width && y >= 0 && y < height {
				for w := 0; w < 3; w++ {
					xx := x + w
					if xx < width {
						img.Set(xx, y, c)
					}
				}
			}
		}
	}
	for j := 0; j < 8; j++ {
		cx := (j*47 + 30) % width
		cy := (j*31 + 20) % height
		r := 12 + j*2
		drawCircleOutline(img, cx, cy, r, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	return buf.Bytes()
}

func drawCircleOutline(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for a := 0; a < 360; a++ {
		rad := float64(a) * math.Pi / 180
		for w := 0; w < 2; w++ {
			x := cx + int(float64(r+w)*math.Cos(rad))
			y := cy + int(float64(r+w)*math.Sin(rad))
			if x >= 0 && y >= 0 && x < img.Bounds().Dx() && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}
