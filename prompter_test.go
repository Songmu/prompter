package prompter

import (
	"fmt"
	"os"
	"regexp"
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
			want:       "y",
			wantStderr: "Continue (y/n) [y]: # Enter `y` or `n`\nContinue (y/n) [y]: ",
		},
		{
			name:       "invalid pattern",
			prompter:   Prompter{Message: "Name", Regexp: regexp.MustCompile(`^[a-z]+$`), Default: "guest"},
			input:      "123\n",
			want:       "guest",
			wantStderr: "Name [guest]: # Answer should match /^[a-z]+$/\nName [guest]: ",
		},
		{
			name:       "password newline",
			prompter:   Prompter{Message: "Password", NoEcho: true},
			wantStderr: "Password: \n",
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

			if got := tt.prompter.Prompt(); got != tt.want {
				t.Errorf("Prompt() = %q, want %q", got, tt.want)
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
