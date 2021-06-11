package gfx

import (
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
)

// CubeMap wraps an OpenGL texture.
type CubeMap struct {
	id        uint32
	width     int32
	layers    int32
	format    uint32
	alignment int32
	texelSize int32
}

// NewCubeMap creates a CubeMap object that wraps the OpenGL texture functions.
// For alignment, see documentation for glPixelStorei.
// Format specifies the memory format of the data.
func NewCubeMap(width, layers int32, data []byte, format int, alignment int32, texelSize int32) (CubeMap, error) {
	t := CubeMap{
		width:     width,
		layers:    layers,
		format:    uint32(format),
		alignment: alignment,
		texelSize: texelSize,
	}
	gl.GenTextures(1, &t.id)
	t.Bind()
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, t.alignment)
	gl.TexImage3D(gl.TEXTURE_CUBE_MAP_ARRAY, 0, int32(format), width, width, layers*6, 0, uint32(format), gl.UNSIGNED_BYTE, unsafe.Pointer(nil))
	for l := int32(0); l < layers; l++ {
		var faceBytes []byte
		for i := int32(0); i < 6; i++ {
			for j := int32(0); j < width; j++ {
				start := (j*6 + i + l*6*width) * 4 * width
				end := start + width*4
				faceBytes = append(faceBytes, data[start:end]...)
			}
		}
		gl.TexSubImage3D(gl.TEXTURE_CUBE_MAP_ARRAY, 0, 0, 0, l*6, width, width, 6, uint32(format), gl.UNSIGNED_BYTE, unsafe.Pointer(&faceBytes[0]))
	}
	// TODO call for every face or once per cubemap??
	gl.GenerateMipmap(gl.TEXTURE_CUBE_MAP_ARRAY)
	t.Unbind()
	return t, nil
}

// SetParameter sets the given parameter for the texture.
func (t CubeMap) SetParameter(paramName uint32, param int32) {
	t.Bind()
	// TODO all 6 ?
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, paramName, param)
	t.Unbind()
}

// Bind sets this texture as the current texture.
func (t CubeMap) Bind() {
	gl.BindTexture(gl.TEXTURE_CUBE_MAP_ARRAY, t.id)
}

// Unbind unsets the current texture.
func (t CubeMap) Unbind() {
	gl.BindTexture(gl.TEXTURE_CUBE_MAP_ARRAY, 0)
}

// GetWidth returns the width of the texture.
func (t CubeMap) GetWidth() int32 {
	return t.width
}

// GetHeight returns the height of the texture.
func (t CubeMap) GetLayers() int32 {
	return t.layers
}

// Destroy frees external resources.
func (t CubeMap) Destroy() {
	gl.DeleteTextures(1, &t.id)
}
