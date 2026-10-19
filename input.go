package prompter

import (
	"bufio"
	"io"
	"strings"
)

func readLine(r io.Reader) (string, error) {
	var input strings.Builder
	var b [1]byte
	// Avoid reading ahead into answers for later prompts or other stdin users.
	for input.Len() < bufio.MaxScanTokenSize {
		n, err := r.Read(b[:])
		if err != nil && err != io.EOF {
			return "", err
		}
		if n > 0 {
			if b[0] == '\n' {
				return strings.TrimRight(input.String(), "\r"), nil
			}
			input.WriteByte(b[0])
		}
		if err == io.EOF {
			if input.Len() == 0 {
				return "", io.EOF
			}
			return strings.TrimRight(input.String(), "\r"), nil
		}
	}
	return "", bufio.ErrTooLong
}
