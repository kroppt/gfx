package gfx

import (
	"github.com/go-gl/gl/v2.1/gl"
)

// FrameBuffer wraps an OpenGL framebuffer.
type FrameBuffer struct {
	id  uint32
	tex Texture
}

// ErrFrameBuffer indicates that a program failed to link.
const ErrFrameBuffer constErr = "incomplete framebuffer"

// NewFrameBuffer creates an FBO of the specified size that renders to
// a texture.
func NewFrameBuffer(width, height int32) (FrameBuffer, error) {
	var fb FrameBuffer
	var err error
	gl.GenFramebuffers(1, &fb.id)
	fb.Bind()
	bufs := uint32(gl.COLOR_ATTACHMENT0)
	gl.DrawBuffers(1, &bufs)

	fb.tex, err = NewTexture(width, height, nil, gl.RGBA, 4, 4)
	if err != nil {
		fb.Unbind()
		return FrameBuffer{}, err
	}
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, fb.tex.id, 0)

	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	fb.Unbind()
	if status != gl.FRAMEBUFFER_COMPLETE {
		return FrameBuffer{}, ErrFrameBuffer
	}
	return fb, nil
}

// GetTexture returns the texture associated with the frame buffer.
func (fb FrameBuffer) GetTexture() Texture {
	return fb.tex
}

// Bind sets this framebuffer to the current framebuffer.
func (fb FrameBuffer) Bind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.id)
}

// Unbind unsets the current framebuffer.
func (fb FrameBuffer) Unbind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

// Destroy frees external resources.
func (fb FrameBuffer) Destroy() {
	gl.DeleteFramebuffers(1, &fb.id)
}
