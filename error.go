package x

type xError struct{}
type writeError struct{}

func (xe *xError) Error() string {
	return "the x error"
}

func (we *writeError) Error() string {
	return "the x writeError"
}
