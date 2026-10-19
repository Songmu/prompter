package prompter

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"testing/iotest"
)

func TestReadLineKeepsFollowingInput(t *testing.T) {
	r := strings.NewReader("first\r\nsecond\nlast")
	for _, want := range []string{"first", "second", "last"} {
		got, err := readLine(r)
		if got != want || err != nil {
			t.Fatalf("readLine() = (%q, %v), want (%q, nil)", got, err, want)
		}
	}
	if got, err := readLine(r); got != "" || !errors.Is(err, io.EOF) {
		t.Errorf("readLine() = (%q, %v), want empty answer and EOF", got, err)
	}
}

func TestReadLine(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  string
		err   error
	}{
		{name: "empty line", input: "\n"},
		{name: "whitespace", input: "  hello  \n", want: "  hello  "},
		{name: "CRLF", input: "hello\r\n", want: "hello"},
		{name: "trailing carriage returns", input: "hello\r\r\n", want: "hello"},
		{
			name:  "maximum line",
			input: strings.Repeat("x", bufio.MaxScanTokenSize-1) + "\n",
			want:  strings.Repeat("x", bufio.MaxScanTokenSize-1),
		},
		{
			name:  "oversized line",
			input: strings.Repeat("x", bufio.MaxScanTokenSize) + "\n",
			err:   bufio.ErrTooLong,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readLine(strings.NewReader(tt.input))
			if got != tt.want || !errors.Is(err, tt.err) {
				t.Errorf("readLine() returned %d bytes and %v, want %d bytes and %v", len(got), err, len(tt.want), tt.err)
			}
		})
	}
}

func TestReadLineReadErrors(t *testing.T) {
	wantErr := errors.New("read failed")
	for _, tt := range []struct {
		name   string
		reader io.Reader
	}{
		{name: "before input", reader: iotest.ErrReader(wantErr)},
		{name: "after partial input", reader: io.MultiReader(strings.NewReader("partial"), iotest.ErrReader(wantErr))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readLine(tt.reader)
			if got != "" || !errors.Is(err, wantErr) {
				t.Errorf("readLine() = (%q, %v), want empty answer and read error", got, err)
			}
		})
	}
}

func TestPromptsKeepFollowingInput(t *testing.T) {
	promptStreams(t, "Alice\nBob\nremaining\n")
	if got := Prompt("First", "fallback-1"); got != "Alice" {
		t.Fatalf("first Prompt() = %q, want Alice", got)
	}
	p := &Prompter{Message: "Second", Default: "fallback-2"}
	if got, err := p.PromptWithError(); got != "Bob" || err != nil {
		t.Fatalf("second PromptWithError() = (%q, %v), want (Bob, nil)", got, err)
	}
	remaining, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatal(err)
	}
	if string(remaining) != "remaining\n" {
		t.Errorf("remaining input = %q, want remaining line", remaining)
	}
}

func TestPromptRetriesKeepFollowingInput(t *testing.T) {
	stderr := promptStreams(t, "invalid\nn\nremaining\n")
	if got := Choose("Continue", []string{"y", "n"}, "y"); got != "n" {
		t.Fatalf("Choose() = %q, want n", got)
	}
	output, err := os.ReadFile(stderr.Name())
	if err != nil {
		t.Fatal(err)
	}
	wantStderr := "Continue (y/n) [y]: # Enter `y` or `n`\nContinue (y/n) [y]: "
	if string(output) != wantStderr {
		t.Errorf("stderr = %q, want %q", output, wantStderr)
	}
	remaining, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatal(err)
	}
	if string(remaining) != "remaining\n" {
		t.Errorf("remaining input = %q, want remaining line", remaining)
	}
}

func TestPromptsFromPipe(t *testing.T) {
	promptStreams(t, "")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close() })
	t.Cleanup(func() { w.Close() })
	if _, err := w.WriteString("invalid\nn\nBob\n"); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	if got := Choose("Continue", []string{"y", "n"}, "y"); got != "n" {
		t.Fatalf("Choose() = %q, want n", got)
	}
	if got := Prompt("Name", "fallback"); got != "Bob" {
		t.Errorf("Prompt() = %q, want Bob", got)
	}
}
