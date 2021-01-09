package gfx

type constErr string

func (e constErr) Error() string {
	return string(e)
}
