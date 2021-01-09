package gfx

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
)

// Texture wraps an OpenGL texture.
type Texture struct {
	id        uint32
	width     int32
	height    int32
	format    uint32
	alignment int32
	texelSize int32
}

// NewTextureFromFile creates a new Texture, loading data from fileName
// with the assumption that it is an image that can be converted to RGBA
// (alpha is black for jpegs).
//
// To provide support for loading different image types, blank import the
// respective image/* packages.
func NewTextureFromFile(fileName string) (Texture, error) {
	in, err := os.Open(fileName)
	if err != nil {
		return Texture{}, err
	}
	defer in.Close()

	img, _, err := image.Decode(in)
	if err != nil {
		return Texture{}, err
	}
	// TODO load from underlying arrays directly and correctly format in OpenGL
	// switch img.(type) {
	// case *image.Alpha:
	// case *image.Alpha16:
	// case *image.CMYK:
	// case *image.Gray:
	// case *image.Gray16:
	// case *image.NRGBA:
	// case *image.NRGBA64:
	// case *image.Paletted:
	// case *image.RGBA:
	// case *image.RGBA64:
	// case *image.YCbCr, *image.NYCbCrA, *image.Uniform:
	// 	// no Pix array
	// }
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	data := make([]byte, 0, width*height*4)
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			col := color.NRGBAModel.Convert(img.At(i, j))
			nrgba := col.(color.NRGBA)
			r, g, b, a := nrgba.R, nrgba.G, nrgba.B, nrgba.A
			data = append(data, r, g, b, a)
		}
	}
	t, err := NewTexture(int32(width), int32(height), data, gl.RGBA, 4, 4)
	t.SetParameter(gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	t.SetParameter(gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	return t, err
}

// NewTexture creates a Texture object that wraps the OpenGL texture functions.
// For alignment, see documentation for glPixelStorei.
// Format specifies the memory format of the data.
func NewTexture(width, height int32, data []byte, format int, alignment int32, texelSize int32) (Texture, error) {
	t := Texture{
		width:     width,
		height:    height,
		format:    uint32(format),
		alignment: alignment,
		texelSize: texelSize,
	}
	var ptr unsafe.Pointer
	if data != nil {
		ptr = unsafe.Pointer(&data[0])
	}
	gl.GenTextures(1, &t.id)
	t.Bind()
	// copy pixels to texture
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, t.alignment)
	gl.TexImage2D(gl.TEXTURE_2D, 0, int32(format), width, height, 0, uint32(format), gl.UNSIGNED_BYTE, ptr)
	gl.GenerateMipmap(gl.TEXTURE_2D)
	t.Unbind()

	return t, nil
}

// SetParameter sets the given parameter for the texture.
func (t Texture) SetParameter(paramName uint32, param int32) {
	t.Bind()
	gl.TexParameteri(gl.TEXTURE_2D, paramName, param)
	t.Unbind()
}

// ErrCoordOutOfRange indicates that given coordinates are out of range.
const ErrCoordOutOfRange constErr = "coordinates out of range"

// SetPixelArea sets the area of a texture to the given data.
func (t Texture) SetPixelArea(r Rect, d []byte, genMipmap bool) error {
	if r.X < 0 || r.Y < 0 || r.X >= t.width || r.Y >= t.height {
		return fmt.Errorf("SetPixelArea(%v): %w", r, ErrCoordOutOfRange)
	}
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, t.alignment)
	gl.TextureSubImage2D(t.id, 0, r.X, r.Y, r.W, r.H, t.format, gl.UNSIGNED_BYTE, unsafe.Pointer(&d[0]))
	if genMipmap {
		t.Bind()
		gl.GenerateMipmap(gl.TEXTURE_2D)
		t.Unbind()
	}
	return nil
}

// SetPixel sets the texture at the given point to the given byte.
func (t Texture) SetPixel(p Point, b byte, genMipmap bool) error {
	return t.SetPixelArea(Rect{X: p.X, Y: p.Y, W: 1, H: 1}, []byte{b}, genMipmap)
}

// GetData returns a byte slice of all the texture data
func (t Texture) GetData() []byte {
	// TODO do this in batches/stream to avoid memory limitations
	var data = make([]byte, t.width*t.height*t.texelSize)
	t.Bind()
	gl.PixelStorei(gl.PACK_ALIGNMENT, t.alignment)
	gl.GetTexImage(gl.TEXTURE_2D, 0, t.format, gl.UNSIGNED_BYTE, unsafe.Pointer(&data[0]))
	t.Unbind()
	return data
}

// GetSubData returns a portion of the texture data specified by the given Rect.
func (t Texture) GetSubData(r Rect) []byte {
	// TODO do this in batches/stream to avoid memory limitations
	var data = make([]byte, r.W*r.H*t.texelSize)
	gl.PixelStorei(gl.PACK_ALIGNMENT, t.alignment)
	gl.GetTextureSubImage(t.id, 0, r.X, r.Y, 0, r.W, r.H, 1, t.format, gl.UNSIGNED_BYTE, r.W*r.H*t.texelSize, unsafe.Pointer(&data[0]))
	return data
}

// Bind sets this texture as the current texture.
func (t Texture) Bind() {
	gl.BindTexture(gl.TEXTURE_2D, t.id)
}

// Unbind unsets the current texture.
func (t Texture) Unbind() {
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// GetWidth returns the width of the texture.
func (t Texture) GetWidth() int32 {
	return t.width
}

// GetHeight returns the height of the texture.
func (t Texture) GetHeight() int32 {
	return t.height
}

// Destroy frees external resources.
func (t Texture) Destroy() {
	gl.DeleteTextures(1, &t.id)
}
