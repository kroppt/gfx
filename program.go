package gfx

import (
	"fmt"

	"github.com/go-gl/gl/v2.1/gl"
)

// Program wraps an OpenGL program.
type Program struct {
	id uint32
}

// ErrProgramLink indicates that a program failed to link.
const ErrProgramLink constErr = "failed to link program"

// NewProgram compiles a vertex and fragment shader, attaches them to a new
// shader program and returns its ID.
func NewProgram(shaders ...Shader) (Program, error) {
	prog := gl.CreateProgram()
	for _, shader := range shaders {
		gl.AttachShader(prog, shader.id)
	}
	gl.LinkProgram(prog)

	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLength)

		log := string(make([]byte, logLength+1))
		gl.GetProgramInfoLog(prog, logLength, nil, gl.Str(log))

		return Program{}, fmt.Errorf("%w: %v", ErrProgramLink, log)
	}

	return Program{prog}, nil
}

// ErrInvalidName indicates that a given name was invalid.
const ErrInvalidName constErr = "invalid name"

// ErrInvalidNumberArgs indicates an invalid number of arguments were given.
const ErrInvalidNumberArgs constErr = "invalid number of arguments"

// UploadUniform uploads float32 data in the given uniform variable
// belonging to the given program ID.
//
// Possible errors are ErrInvalidName and ErrInvalidNumberArgs.
func (p Program) UploadUniform(uniformName string, data ...float32) error {
	uniformID := gl.GetUniformLocation(p.id, &[]byte(uniformName + "\x00")[0])
	if uniformID == -1 {
		return fmt.Errorf("%w: \"%v\"", ErrInvalidName, uniformName)
	}
	gl.UseProgram(p.id)
	switch len(data) {
	case 1:
		gl.Uniform1f(uniformID, data[0])
	case 2:
		gl.Uniform2f(uniformID, data[0], data[1])
	case 3:
		gl.Uniform3f(uniformID, data[0], data[1], data[2])
	case 4:
		gl.Uniform4f(uniformID, data[0], data[1], data[2], data[3])
	default:
		return fmt.Errorf("%w: %v (max 4)", ErrInvalidNumberArgs, len(data))
	}
	gl.UseProgram(0)
	return nil
}

// UploadUniformi uploads int32 data in the given uniform variable belonging to
// the given program ID.
//
// Possible errors are ErrInvalidName and ErrInvalidNumberArgs.
func (p Program) UploadUniformi(uniformName string, data ...int32) error {
	uniformID := gl.GetUniformLocation(p.id, &[]byte(uniformName + "\x00")[0])
	if uniformID == -1 {
		return fmt.Errorf("%w: \"%v\"", ErrInvalidName, uniformName)
	}
	gl.UseProgram(p.id)
	switch len(data) {
	case 1:
		gl.Uniform1i(uniformID, data[0])
	case 2:
		gl.Uniform2i(uniformID, data[0], data[1])
	case 3:
		gl.Uniform3i(uniformID, data[0], data[1], data[2])
	case 4:
		gl.Uniform4i(uniformID, data[0], data[1], data[2], data[3])
	default:
		return fmt.Errorf("%w: %v (max 4)", ErrInvalidNumberArgs, len(data))
	}
	gl.UseProgram(0)
	return nil
}

// UploadUniformui uploads uint32 data in the given uniform variable belonging
// to the given program ID.
//
// Possible errors are ErrInvalidName and ErrInvalidNumberArgs.
func (p Program) UploadUniformui(uniformName string, data ...uint32) error {
	uniformID := gl.GetUniformLocation(p.id, &[]byte(uniformName + "\x00")[0])
	if uniformID == -1 {
		return fmt.Errorf("%w: \"%v\"", ErrInvalidName, uniformName)
	}
	gl.UseProgram(p.id)
	switch len(data) {
	case 1:
		gl.Uniform1uiEXT(uniformID, data[0])
	case 2:
		gl.Uniform2uiEXT(uniformID, data[0], data[1])
	case 3:
		gl.Uniform3uiEXT(uniformID, data[0], data[1], data[2])
	case 4:
		gl.Uniform4uiEXT(uniformID, data[0], data[1], data[2], data[3])
	default:
		return fmt.Errorf("%w: %v (max 4)", ErrInvalidNumberArgs, len(data))
	}
	gl.UseProgram(0)
	return nil
}

// UploadUniformMat4 uploads matrix data in the given uniform variable belonging
// to the given program ID.
//
// Possible errors are ErrInvalidName.
func (p Program) UploadUniformMat4(uniformName string, data [16]float32) error {
	uniformID := gl.GetUniformLocation(p.id, &[]byte(uniformName + "\x00")[0])
	if uniformID == -1 {
		return fmt.Errorf("%w: \"%v\"", ErrInvalidName, uniformName)
	}
	gl.UseProgram(p.id)
	gl.UniformMatrix4fv(uniformID, 1, false, &data[0])
	gl.UseProgram(0)
	return nil
}

// Bind sets the program to the current program.
func (p Program) Bind() {
	gl.UseProgram(p.id)
}

// Unbind unsets the current program.
func (p Program) Unbind() {
	gl.UseProgram(0)
}

// Destroy frees external resources.
func (p Program) Destroy() {
	gl.DeleteProgram(p.id)
}
