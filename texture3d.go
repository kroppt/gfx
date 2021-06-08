package gfx

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
)

// Texture3D wraps an OpenGL texture.
type Texture3D struct {
	id        uint32
	width     int32
	height    int32
	depth     int32
	format    uint32
	alignment int32
	texelSize int32
}

func NewTexture3D(width, height, depth int32, data []byte, format int, alignment int32, texelSize int32) (Texture3D, error) {
	t := Texture3D{
		width:     width,
		height:    height,
		depth:     depth,
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
	gl.TexImage3D(gl.TEXTURE_3D, 0, int32(format), width, height, depth, 0, uint32(format), gl.UNSIGNED_BYTE, ptr)
	gl.GenerateMipmap(gl.TEXTURE_3D)
	t.Unbind()

	return t, nil
}

// SetParameter sets the given parameter for the texture.
func (t Texture3D) SetParameter(paramName uint32, param int32) {
	t.Bind()
	gl.TexParameteri(gl.TEXTURE_3D, paramName, param)
	t.Unbind()
}

// SetPixelArea sets the area of a texture to the given data.
func (t Texture3D) SetPixelArea(x, y, z, w, h, depth int32, d []byte, genMipmap bool) error {
	if x < 0 || y < 0 || z < 0 || x >= t.width || y >= t.height || z >= t.depth {
		return fmt.Errorf("SetPixelArea(%v %v %v %v %v %v): %w", x, y, z, w, h, depth, ErrCoordOutOfRange)
	}
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, t.alignment)
	gl.TextureSubImage3D(t.id, 0, x, y, z, w, h, depth, t.format, gl.UNSIGNED_BYTE, unsafe.Pointer(&d[0]))
	if genMipmap {
		t.Bind()
		gl.GenerateMipmap(gl.TEXTURE_3D)
		t.Unbind()
	}
	return nil
}

// SetPixel sets the texture at the given point to the given data.
func (t Texture3D) SetPixel(p Point3D, d []byte, genMipmap bool) error {
	return t.SetPixelArea(p.X, p.Y, p.Z, 1, 1, 1, d, genMipmap)
}

// GetData returns a byte slice of all the texture data
func (t Texture3D) GetData() []byte {
	// TODO do this in batches/stream to avoid memory limitations
	var data = make([]byte, t.width*t.height*t.depth*t.texelSize)
	t.Bind()
	gl.PixelStorei(gl.PACK_ALIGNMENT, t.alignment)
	gl.GetTexImage(gl.TEXTURE_3D, 0, t.format, gl.UNSIGNED_BYTE, unsafe.Pointer(&data[0]))
	t.Unbind()
	return data
}

// Bind sets this texture as the current texture.
func (t Texture3D) Bind() {
	gl.BindTexture(gl.TEXTURE_3D, t.id)
}

// Unbind unsets the current texture.
func (t Texture3D) Unbind() {
	gl.BindTexture(gl.TEXTURE_3D, 0)
}

// GetWidth returns the width of the texture.
func (t Texture3D) GetWidth() int32 {
	return t.width
}

// GetHeight returns the height of the texture.
func (t Texture3D) GetHeight() int32 {
	return t.height
}

// Destroy frees external resources.
func (t Texture3D) Destroy() {
	gl.DeleteTextures(1, &t.id)
}
