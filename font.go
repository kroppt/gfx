package gfx

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"unicode"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

const minASCII = 32

func int26_6ToFloat32(x fixed.Int26_6) float32 {
	top := float32(x >> 6)
	bottom := float32(x&0x3F) / 64.0
	return top + bottom
}

// AlignV is used for the positioning of elements vertically.
type AlignV int

const (
	// AlignAbove puts the top side at the y coordinate.
	AlignAbove AlignV = iota - 1
	// AlignMiddle puts the top and bottom sides equidistant from the middle.
	AlignMiddle
	// AlignBelow puts the bottom side on the y coordinate.
	AlignBelow
)

// AlignH is used for the positioning of elements horizontally.
type AlignH int

const (
	// AlignLeft puts the left side on the x coordinate.
	AlignLeft AlignH = iota - 1
	//AlignCenter puts the left and right sides equidistant from the center.
	AlignCenter
	// AlignRight puts the right side at the x coordinate.
	AlignRight
)

// Align holds vertical and horizontal alignments.
type Align struct {
	V AlignV
	H AlignH
}

type runeInfo struct {
	row      int32
	width    int32
	height   int32
	bearingX float32
	bearingY float32
	advance  float32
}

// FontInfo represents a loaded font.
type FontInfo struct {
	texture Texture    // texture of cached glyph data
	runeMap []runeInfo // map of character-specific spacing info
	metrics metrics
}

type metrics struct {
	Height     float32
	Ascent     float32
	Descent    float32
	XHeight    float32
	CapHeight  float32
	CaretSlope image.Point
}

// GetTexture returns the font's OpenGL texture.
func (font *FontInfo) GetTexture() Texture {
	return font.texture
}

// MapString turns each character in the string into a pair of
// (x,y,s,t)-vertex triangles using glyph information from a
// pre-loaded font. The vertex info is returned as []float32.
func (font *FontInfo) MapString(str string, pos Point, align Align) []float32 {
	// 2 triangles per rune, 3 vertices per triangle, 4 float32's per vertex (x,y,s,t)
	buffer := make([]float32, 0, len(str)*24)
	// get glyph information for alignment
	var strWidth float32
	for _, r := range str {
		info := font.runeMap[r-minASCII]
		strWidth += info.advance
	}
	// adjust strWidth if last rune's width + bearingX > advance
	lastInfo := font.runeMap[str[len(str)-1]-minASCII]
	if float32(lastInfo.width)+lastInfo.bearingX > lastInfo.advance {
		strWidth += (float32(lastInfo.width) + lastInfo.bearingX - lastInfo.advance)
	}

	w2 := float64(strWidth) / 2.0
	offx := int32(-w2 - float64(align.H)*w2)
	var offy float32
	switch align.V {
	case AlignBelow:
		offy = -float32(math.Ceil(float64(font.metrics.Ascent)))
	case AlignMiddle:
		offy = -font.metrics.XHeight / 2
	case AlignAbove:
		offy = float32(math.Ceil(float64(font.metrics.Descent)))
	}
	// offset origin to account for alignment

	type pointF32 struct {
		x float32
		y float32
	}

	origin := pointF32{float32(pos.X + offx), float32(pos.Y) + offy}
	for _, r := range str {
		info := font.runeMap[r-minASCII]

		// calculate x,y position coordinates - use bottom left as (0,0); shader converts for you
		posTL := pointF32{origin.x + info.bearingX, origin.y + (float32(info.height) - info.bearingY)}
		posTR := pointF32{posTL.x + float32(info.width), posTL.y}
		posBL := pointF32{posTL.x, origin.y - info.bearingY}
		posBR := pointF32{posTR.x, posBL.y}
		// calculate s,t texture coordinates - use top left as (0,0); shader converts for you
		texTL := pointF32{0, float32(info.row)}
		texTR := pointF32{float32(info.width), texTL.y}
		texBL := pointF32{texTL.x, texTL.y + float32(info.height)}
		texBR := pointF32{texTR.x, texBL.y}
		// create 2 triangles
		triangles := []float32{
			posBL.x, posBL.y, texBL.x, texBL.y, // bottom-left
			posTL.x, posTL.y, texTL.x, texTL.y, // top-left
			posTR.x, posTR.y, texTR.x, texTR.y, // top-right

			posBL.x, posBL.y, texBL.x, texBL.y, // bottom-left
			posTR.x, posTR.y, texTR.x, texTR.y, // top-right
			posBR.x, posBR.y, texBR.x, texBR.y, // bottom-right
		}
		buffer = append(buffer, triangles...)

		origin.x += info.advance
	}

	return buffer
}

type fontKey struct {
	fontName string
	fontSize int32
}

// fontMap caches previously loaded fonts
var fontMap map[fontKey]FontInfo

// ErrNoFontGlyph indicates the given font does not contain the given glyph.
var ErrNoFontGlyph error = fmt.Errorf("font does not contain given glyph")

// LoadFontTexture caches all of the glyph pixel data in an OpenGL texture for
// a given font at a given size. It returns an Info struct populated with the
// OpenGL ID for this texture, metrics, and an array containing glyph spacing info.
func LoadFontTexture(fontName string, fontSize int32) (*FontInfo, error) {
	if fontMap == nil {
		fontMap = make(map[fontKey]FontInfo)
	}
	if val, ok := fontMap[fontKey{fontName, fontSize}]; ok {
		return &val, nil
	}

	var err error
	var fontBytes []byte
	var ttfFont *truetype.Font
	if fontBytes, err = ioutil.ReadFile(fontName); err != nil {
		return nil, err
	}
	if ttfFont, err = truetype.Parse(fontBytes); err != nil {
		return nil, err
	}
	face := truetype.NewFace(ttfFont, &truetype.Options{Size: float64(fontSize)})

	var sfntFont *sfnt.Font
	if fontBytes, err = ioutil.ReadFile(fontName); err != nil {
		return nil, err
	}
	if sfntFont, err = sfnt.Parse(fontBytes); err != nil {
		return nil, err
	}

	var runeMap [unicode.MaxASCII - minASCII]runeInfo
	var glyphBytes []byte
	var currentIndex int32
	for i := minASCII; i < unicode.MaxASCII; i++ {
		c := rune(i)

		roundedRect, mask, maskp, advance, okGlyph := face.Glyph(fixed.Point26_6{X: 0, Y: 0}, c)
		if !okGlyph {
			return nil, fmt.Errorf("LoadFontTexture(\"%v\", %v) glyph '%v': %w", fontName, fontSize, c, ErrNoFontGlyph)
		}
		accurateRect, _, okBounds := face.GlyphBounds(c)
		glyph, okCast := mask.(*image.Alpha)
		if !okBounds || !okCast {
			return nil, fmt.Errorf("LoadFontTexture(\"%v\", %v) glyph '%v': %w", fontName, fontSize, c, ErrNoFontGlyph)
		}

		runeMap[i-minASCII] = runeInfo{
			row:      currentIndex,
			width:    int32(roundedRect.Dx()),
			height:   int32(roundedRect.Dy()),
			bearingX: float32(math.Round(float64(accurateRect.Min.X.Ceil()))),
			bearingY: float32(accurateRect.Max.Y.Ceil()),
			advance:  float32(math.Round(float64(int26_6ToFloat32(advance)))),
		}
		// alternatively, upload entire glyph cache into OpenGL texture
		// ... but this doesnt take that long and cuts texture size by 95%
		for row := 0; row < roundedRect.Dy(); row++ {
			beg := (maskp.Y + row) * glyph.Stride
			end := (maskp.Y + row + 1) * glyph.Stride
			glyphBytes = append(glyphBytes, glyph.Pix[beg:end]...)
			currentIndex++
		}
	}

	_, mask, _, _, aOK := face.Glyph(fixed.Point26_6{X: 0, Y: 0}, 'A')
	if !aOK {
		return nil, fmt.Errorf("LoadFontTexture(\"%v\", %v) glyph 'A': %w", fontName, fontSize, ErrNoFontGlyph)
	}

	glyph, _ := mask.(*image.Alpha)
	texWidth := int32(glyph.Stride)
	texHeight := int32(len(glyphBytes) / glyph.Stride)

	// pass glyphBytes to OpenGL texture
	fontTexture, err := NewTexture(texWidth, texHeight, glyphBytes, gl.RED, 1, 1)
	if err != nil {
		return nil, err
	}
	fontTexture.SetParameter(gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	fontTexture.SetParameter(gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	otfFace, err := opentype.NewFace(sfntFont, &opentype.FaceOptions{
		Size:    float64(fontSize),
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return nil, err
	}
	otfMetrics := otfFace.Metrics()
	metrics := metrics{
		Height:     int26_6ToFloat32(otfMetrics.Height),
		Ascent:     int26_6ToFloat32(otfMetrics.Ascent),
		Descent:    int26_6ToFloat32(otfMetrics.Descent),
		XHeight:    int26_6ToFloat32(otfMetrics.XHeight),
		CapHeight:  int26_6ToFloat32(otfMetrics.CapHeight),
		CaretSlope: otfMetrics.CaretSlope,
	}

	InfoLoaded := FontInfo{fontTexture, runeMap[:], metrics}
	fontMap[fontKey{fontName, fontSize}] = InfoLoaded
	return &InfoLoaded, nil
}

// CalcStringDims returns the width and height of a string
func (font *FontInfo) CalcStringDims(str string) (float64, float64) {
	var strWidth, largestBearingY float32
	for _, r := range str {
		info := font.runeMap[r-minASCII]
		if info.bearingY > largestBearingY {
			largestBearingY = info.bearingY

		}
		strWidth += info.advance
	}
	// adjust strWidth if last rune's width + bearingX > advance
	lastInfo := font.runeMap[str[len(str)-1]-minASCII]
	if float32(lastInfo.width)+lastInfo.bearingX > lastInfo.advance {
		strWidth += (float32(lastInfo.width) + lastInfo.bearingX - lastInfo.advance)
	}

	return float64(strWidth), float64(font.metrics.Height)
}

// WriteFontToFile saves an image of all font characters to fileName.
func (font *FontInfo) WriteFontToFile(fileName string) error {
	width := int(font.texture.GetWidth())
	height := int(font.texture.GetHeight())
	alphaImg := image.NewAlpha(image.Rect(0, 0, width, height))
	outImg := image.NewNRGBA(image.Rect(0, 0, width, height))
	alphaImg.Pix = font.texture.GetData()
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			col := color.NRGBAModel.Convert(alphaImg.At(i, j))
			alpha := col.(color.NRGBA).A
			newCol := color.NRGBA{alpha, alpha, alpha, 255}
			outImg.Set(i, j, newCol)
		}
	}
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	if err = png.Encode(file, outImg); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return nil
}

// func writeRuneToFile(fileName string, mask image.Image, maskp image.Point, rec image.Rectangle) error {
// 	if alpha, ok := mask.(*image.Alpha); ok {
// 		diff := image.Point{rec.Dx(), rec.Dy()}
// 		tofile := alpha.SubImage(image.Rectangle{maskp, maskp.Add(diff)})
// 		if f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755); err != nil {
// 			err = png.Encode(f, tofile)
// 			return err
// 		}
// 	}
// 	return nil
// }

// func printRune(mask image.Image, maskp image.Point, rec image.Rectangle) {
// 	var alpha *image.Alpha
// 	var ok bool
// 	if alpha, ok = mask.(*image.Alpha); !ok {
// 		// log.Warn("printRune image not Alpha")
// 		return
// 	}
// 	out := "PrintRune\n"
// 	for y := maskp.Y; y < maskp.Y+rec.Dy(); y++ {
// 		for x := maskp.X; x < maskp.X+rec.Dx(); x++ {
// 			if _, _, _, a := alpha.At(x, y).RGBA(); a > 0 {
// 				out += fmt.Sprintf("%02x ", byte(a))
// 			} else {
// 				out += ".  "
// 			}
// 		}
// 		out += "\n"
// 	}
// 	// log.Debug(out)
// }
