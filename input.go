package prompter

import (
	"bufio"
	"io"
	"strings"
)

func readLine(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	ok := scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if !ok {
		return "", io.EOF
	}
	return strings.TrimRight(scanner.Text(), "\r\n"), nil
}
