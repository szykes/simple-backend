package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/szykes/simple-backend/errors"
)

func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	nRead, err := rand.Read(b)
	if err != nil {
		return nil, errors.Wrap(err, "rand bytes: failed to read bytes")
	}
	if nRead < n {
		return nil, fmt.Errorf("rand bytes: didn't read enough random bytes")
	}
	return b, nil
}

func String(n int) (string, error) {
	b, err := Bytes(n)
	if err != nil {
		return "", errors.Wrap(err, "rand string: failed to get the given length of bytes: %v", n)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
