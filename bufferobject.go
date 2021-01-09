package gfx

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
)

// BufferObject wraps an OpenGL buffer.
type BufferObject struct {
	id        uint32
	sizeBytes uint32
}

// NewBufferObject returns a new buffer object.
func NewBufferObject() *BufferObject {
	var bo BufferObject
	gl.GenBuffers(1, &bo.id)
	bo.sizeBytes = 0
	return &bo
}

// GetSizeBytes returns the data store's size in bytes.
func (bo *BufferObject) GetSizeBytes() uint32 {
	return bo.sizeBytes
}

// BufferData Creates and initializes the buffer data store.
func (bo *BufferObject) BufferData(target uint32, sizeBytes uint32, ptr unsafe.Pointer, usage uint32) {
	bo.sizeBytes = sizeBytes
	bo.Bind(target)
	gl.BufferData(target, int(sizeBytes), ptr, usage)
	bo.Unbind(target)
}

// ErrOutOfBounds indicates that the input was out of bounds.
const ErrOutOfBounds constErr = "out of bounds"

// BufferSubData updates a portion of the buffer data store.
func (bo *BufferObject) BufferSubData(target, offset, sizeBytes uint32, ptr unsafe.Pointer) error {
	// gl.BufferData acts like malloc, while gl.BufferSubData acts like memcpy
	// BufferSubData can only modify a range of the existing size
	if offset+sizeBytes > bo.sizeBytes {
		return fmt.Errorf("%w: %v > %v", ErrOutOfBounds, offset+sizeBytes, bo.sizeBytes)
	}
	bo.Bind(target)
	gl.BufferSubData(target, int(offset), int(sizeBytes), ptr)
	bo.Unbind(target)
	return nil
}

// GetBufferSubData returns a subset of the buffer data store.
func (bo *BufferObject) GetBufferSubData(target, offset, sizeBytes uint32, ptr unsafe.Pointer) {
	bo.Bind(target)
	gl.GetBufferSubData(target, int(offset), int(sizeBytes), ptr)
	bo.Unbind(target)
}

// GetData returns all of the buffer data store.
func (bo *BufferObject) GetData(target uint32, ptr unsafe.Pointer) {
	bo.GetBufferSubData(target, 0, bo.sizeBytes, ptr)
}

// Bind sets the current buffer.
func (bo *BufferObject) Bind(target uint32) {
	gl.BindBuffer(target, bo.id)
}

// Unbind unsets the current buffer.
func (bo *BufferObject) Unbind(target uint32) {
	gl.BindBuffer(target, 0)
}

// BindBufferBase sets the current
func (bo *BufferObject) BindBufferBase(target, binding uint32) {
	gl.BindBufferBase(target, binding, bo.id)
}

// Destroy frees external resources.
func (bo *BufferObject) Destroy() {
	gl.DeleteBuffers(1, &bo.id)
	bo.id = 0
	bo.sizeBytes = 0
}
