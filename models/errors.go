package models

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/szykes/simple-backend/errors"
)

var (
	ErrNotFound   = errors.New("no resource is found")
	ErrEmailTaken = errors.New("email address is already in use")
	ErrPwMismatch = errors.New("mismatching password")
)

type FileError struct {
	Issue string
}

func (f FileError) Error() string {
	return fmt.Sprintf("invalid file: %v", f.Issue)
}

func checkContentType(r io.ReadSeeker, allowedTypes []string) error {
	testBytes := make([]byte, 512)
	_, err := r.Read(testBytes)
	if err != nil {
		return errors.Wrap(err, "checking content type")
	}

	_, err = r.Seek(io.SeekStart, 0)
	if err != nil {
		return errors.Wrap(err, "checking content type")
	}

	contentType := http.DetectContentType(testBytes)
	for _, t := range allowedTypes {
		if contentType == t {
			return nil
		}
	}
	return FileError{
		Issue: fmt.Sprintf("invalid content type: %v", contentType),
	}
}

func checkExtension(filename string, allowedExtensions []string) error {
	if hasExtension(filename, allowedExtensions) {
		return nil
	}
	return FileError{
		Issue: fmt.Sprintf("invalid extension: %v", filepath.Ext(filename)),
	}
}
