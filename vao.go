package gfx

import (
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
)

// VAO represents a Vertex Array Object.
type VAO struct {
	id         uint32
	vbo        *BufferObject
	mode       uint32
	vertSize   int32
	numAttribs uint32
	layout     []int32
}

// NewVAO creates the structure necessary for efficiently rendering shapes in
// OpenGL. It configures a VAO & VBO pair with a specified mode and vertex
// layout.
// Example mode: gl.TRIANGLES.
// Example vertex layout: (x,y,z, s,t) -> layout = (3, 2).
func NewVAO(mode uint32, layout []int32) *VAO {
	var id uint32
	gl.GenVertexArrays(1, &id)
	vbo := NewBufferObject()
	var vertSize int32
	for _, s := range layout {
		vertSize += s
	}
	configureVAO(id, vbo, layout, vertSize)
	return &VAO{
		id:         id,
		vbo:        vbo,
		mode:       mode,
		vertSize:   vertSize,
		numAttribs: uint32(len(layout)),
		layout:     layout,
	}
}

// configureVAO configures a VAO & VBO pair with a specified vertex layout.
// Example vertex layout: (x,y,z, s,t) -> layout = (3, 2).
func configureVAO(vaoID uint32, vbo *BufferObject, layout []int32, vertSize int32) {
	vbo.Bind(gl.ARRAY_BUFFER)
	gl.BindVertexArray(vaoID)

	// calculate vertex size in bytes
	// ex: (x,y,z,s,t) -> 5*4 = 20 bytes
	vertexStride := vertSize * 4
	var offset int32
	for i := 0; i < len(layout); i++ {
		gl.VertexAttribPointer(uint32(i), layout[i], gl.FLOAT, false, vertexStride, unsafe.Pointer(uintptr(offset*4)))
		offset += layout[i]
	}

	gl.BindVertexArray(0)
	vbo.Unbind(gl.ARRAY_BUFFER)
}

// ErrEmptyData indiciates that the given data is empty.
const ErrEmptyData constErr = "data is empty so cannot be used"

// Load calls buffer data on the current VBO.
// Example usage: gl.STATIC_DRAW.
func (vao *VAO) Load(data []float32, usage uint32) error {
	if len(data) == 0 {
		return ErrEmptyData
	}
	vao.vbo.BufferData(gl.ARRAY_BUFFER, uint32(4*len(data)), gl.Ptr(&data[0]), usage)
	return nil
}

// Draw renders the shapes from previously loaded data.
func (vao *VAO) Draw() {
	if vao.vbo.GetSizeBytes() == 0 {
		return
	}
	var i uint32
	gl.BindVertexArray(vao.id)
	for i = 0; i < vao.numAttribs; i++ {
		gl.EnableVertexAttribArray(i)
	}
	gl.DrawArrays(vao.mode, 0, int32(vao.vbo.GetSizeBytes())/(4*vao.vertSize))
	for i = 0; i < vao.numAttribs; i++ {
		gl.DisableVertexAttribArray(i)
	}
	gl.BindVertexArray(0)
}

// Destroy frees external resources.
func (vao *VAO) Destroy() {
	gl.DeleteVertexArrays(1, &vao.id)
	vao.vbo.Destroy()
	vao.mode = 0
	vao.vbo = nil
	vao.id = 0
}
