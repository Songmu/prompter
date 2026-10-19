package prompter

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestPromptStderr(t *testing.T) {
	tests := []struct {
		name       string
		prompter   Prompter
		input      string
		useDefault string
		want       string
		wantStderr string
		wantError  bool
	}{
		{
			name:       "explicit default",
			prompter:   Prompter{Message: "Name", Default: "guest", UseDefault: true},
			want:       "guest",
			wantStderr: "Name [guest]: ",
		},
		{
			name:       "environment default",
			prompter:   Prompter{Message: "Name", Default: "guest"},
			useDefault: "1",
			want:       "guest",
			wantStderr: "Name [guest]: ",
		},
		{
			name:       "input",
			prompter:   Prompter{Message: "Name"},
			input:      "Alice\n",
			want:       "Alice",
			wantStderr: "Name: ",
		},
		{
			name:       "invalid choice",
			prompter:   Prompter{Message: "Continue", Choices: []string{"y", "n"}, Default: "y"},
			input:      "invalid\n",
			wantStderr: "Continue (y/n) [y]: # Enter `y` or `n`\nContinue (y/n) [y]: ",
			wantError:  true,
		},
		{
			name:       "invalid pattern",
			prompter:   Prompter{Message: "Name", Regexp: regexp.MustCompile(`^[a-z]+$`), Default: "guest"},
			input:      "123\n",
			wantStderr: "Name [guest]: # Answer should match /^[a-z]+$/\nName [guest]: ",
			wantError:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GO_PROMPTER_USE_DEFAULT", tt.useDefault)
			t.Setenv("GO_PROMPTER_USE_PIPE", "1")
			streams := make([]*os.File, 3)
			for i := range streams {
				f, err := os.CreateTemp(t.TempDir(), "stream")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { f.Close() })
				streams[i] = f
			}
			if _, err := streams[0].WriteString(tt.input); err != nil {
				t.Fatal(err)
			}
			if _, err := streams[0].Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			stdin, stdout, stderr := os.Stdin, os.Stdout, os.Stderr
			os.Stdin, os.Stdout, os.Stderr = streams[0], streams[1], streams[2]
			t.Cleanup(func() {
				os.Stdin, os.Stdout, os.Stderr = stdin, stdout, stderr
			})

			got, err := tt.prompter.PromptWithError()
			if got != tt.want {
				t.Errorf("PromptWithError() = %q, want %q", got, tt.want)
			}
			if tt.wantError {
				if !errors.Is(err, io.EOF) {
					t.Errorf("PromptWithError() error = %v, want EOF", err)
				}
			} else if err != nil {
				t.Fatalf("PromptWithError(): %v", err)
			}
			for _, stream := range []struct {
				name string
				file *os.File
				want string
			}{
				{"stdout", os.Stdout, ""},
				{"stderr", os.Stderr, tt.wantStderr},
			} {
				got, err := os.ReadFile(stream.file.Name())
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != stream.want {
					t.Errorf("%s = %q, want %q", stream.name, got, stream.want)
				}
			}
		})
	}
}

func promptStreams(t *testing.T, input string) *os.File {
	t.Helper()
	t.Setenv("GO_PROMPTER_USE_DEFAULT", "")
	t.Setenv("GO_PROMPTER_USE_PIPE", "1")
	stdin, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stdin.Close() })
	if _, err := stdin.WriteString(input); err != nil {
		t.Fatal(err)
	}
	if _, err := stdin.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stderr.Close() })
	oldStdin, oldStderr := os.Stdin, os.Stderr
	os.Stdin, os.Stderr = stdin, stderr
	t.Cleanup(func() {
		os.Stdin, os.Stderr = oldStdin, oldStderr
	})
	return stderr
}

func TestPromptWithErrorReadFailures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		prompter Prompter
		close    bool
		wantErr  error
	}{
		{name: "EOF", wantErr: io.EOF},
		{
			name:     "EOF with default",
			prompter: Prompter{Default: "fallback"},
			wantErr:  io.EOF,
		},
		{
			name:     "EOF with choices",
			prompter: Prompter{Choices: []string{"y", "n"}},
			wantErr:  io.EOF,
		},
		{
			name:     "EOF after invalid input",
			input:    "invalid\n",
			prompter: Prompter{Choices: []string{"y", "n"}, Default: "y"},
			wantErr:  io.EOF,
		},
		{
			name:     "oversized input",
			input:    strings.Repeat("x", bufio.MaxScanTokenSize+1),
			prompter: Prompter{Default: "fallback"},
			wantErr:  bufio.ErrTooLong,
		},
		{
			name:     "closed input",
			prompter: Prompter{Default: "fallback"},
			close:    true,
			wantErr:  os.ErrClosed,
		},
		{
			name:     "password without terminal",
			input:    "secret\n",
			prompter: Prompter{NoEcho: true, Default: "fallback"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stderr := promptStreams(t, tt.input)
			if tt.close {
				if err := os.Stdin.Close(); err != nil {
					t.Fatal(err)
				}
			}
			got, err := tt.prompter.PromptWithError()
			if got != "" {
				t.Errorf("PromptWithError() = %q, want empty answer", got)
			}
			if err == nil || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Errorf("PromptWithError() error = %v, want %v", err, tt.wantErr)
			}
			output, err := os.ReadFile(stderr.Name())
			if err != nil {
				t.Fatal(err)
			}
			if tt.prompter.NoEcho && string(output) != tt.prompter.msg()+"\n" {
				t.Errorf("stderr = %q, want password prompt followed by newline", output)
			}
		})
	}
}

func TestPromptReportsReadError(t *testing.T) {
	stderr := promptStreams(t, "")
	p := &Prompter{Message: "Continue", Choices: []string{"y", "n"}, Default: "y"}
	if got := p.Prompt(); got != "" {
		t.Errorf("Prompt() = %q, want empty answer", got)
	}
	output, err := os.ReadFile(stderr.Name())
	if err != nil {
		t.Fatal(err)
	}
	if want := "Continue (y/n) [y]: # read input: EOF\n"; string(output) != want {
		t.Errorf("stderr = %q, want %q", output, want)
	}
}

func TestConfirmReadErrors(t *testing.T) {
	for _, tt := range []struct {
		name   string
		prompt func(string, bool) bool
	}{
		{name: "YN", prompt: YN},
		{name: "YesNo", prompt: YesNo},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stderr := promptStreams(t, "")
			if tt.prompt("Continue", true) {
				t.Error("EOF must not accept an affirmative default")
			}
			output, err := os.ReadFile(stderr.Name())
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(output), "# read input: EOF\n") {
				t.Errorf("stderr = %q, want EOF report", output)
			}
		})
	}
}

func TestPromptWithErrorAnswers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		prompter Prompter
		want     string
	}{
		{name: "answer", input: "Alice\n", want: "Alice"},
		{name: "answer without newline", input: "Alice", want: "Alice"},
		{name: "empty line", input: "\n", prompter: Prompter{Default: "guest"}, want: "guest"},
		{name: "CRLF", input: "Alice\r\n", want: "Alice"},
		{name: "explicit default", prompter: Prompter{Default: "guest", UseDefault: true}, want: "guest"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptStreams(t, tt.input)
			got, err := tt.prompter.PromptWithError()
			if got != tt.want || err != nil {
				t.Errorf("PromptWithError() = (%q, %v), want (%q, nil)", got, err, tt.want)
			}
		})
	}
}

func TestMsg(t *testing.T) {
	p := &Prompter{
		Choices: []string{"aa", "bb", "cc"},
		Default: "aa",
		Message: "plaase select",
	}

	if p.msg() != "plaase select (aa/bb/cc) [aa]: " {
		t.Errorf("something went wrong")
	}

	if p.errorMsg() != "# Enter `aa`, `bb` or `cc`" {
		t.Errorf("something went wrong")
	}

	if !p.inputIsValid("aa") {
		t.Errorf("something went wrong")
	}

	if p.inputIsValid("AA") {
		t.Errorf("something went wrong")
	}

	input := p.Prompt()
	if input != "aa" {
		fmt.Printf("%s\n", input)
		t.Errorf("something went wrong")
	}
}
