package errors

type publicError struct {
	err error
	msg string
}

func (p publicError) Error() string {
	return p.err.Error()
}

func (p publicError) Public() string {
	return p.msg
}

func (p publicError) Unwrap() error {
	return p.err
}

func Public(err error, msg string) error {
	return publicError{
		err: err,
		msg: msg,
	}
}
