// Package prompter is utility for easy prompting
package prompter

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

// VERSION version of prompter
const VERSION = "0.5.1"

// Prompter is object for prompting
type Prompter struct {
	Message string
	// choices of answer
	Choices    []string
	IgnoreCase bool
	Default    string
	// specify answer pattern by regexp. When both Choices and Regexp are specified, Regexp takes a priority.
	Regexp *regexp.Regexp
	// for passwords and so on.
	NoEcho     bool
	UseDefault bool
	reg        *regexp.Regexp
}

// Prompt displays a prompt and returns the answer.
// On a read error, it reports the error to stderr and returns an empty string.
// Use PromptWithError to handle read errors directly.
func (p *Prompter) Prompt() string {
	input, err := p.PromptWithError()
	if err != nil {
		fmt.Fprintf(os.Stderr, "# %v\n", err)
	}
	return input
}

// PromptWithError displays a prompt and returns the answer or a read error.
// EOF and read errors are not treated as empty answers or replaced by Default.
func (p *Prompter) PromptWithError() (string, error) {
	for {
		fmt.Fprint(os.Stderr, p.msg())
		if p.UseDefault || skip() {
			return p.Default, nil
		}
		var input string
		if p.NoEcho {
			b, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprint(os.Stderr, "\n")
			if err != nil {
				return "", fmt.Errorf("read password: %w", err)
			}
			input = string(b)
		} else {
			var err error
			input, err = readLine(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("read input: %w", err)
			}
		}
		if input == "" {
			input = p.Default
		}
		if p.inputIsValid(input) {
			return input, nil
		}
		fmt.Fprintln(os.Stderr, p.errorMsg())
	}
}

func skip() bool {
	if os.Getenv("GO_PROMPTER_USE_DEFAULT") != "" {
		return true
	}
	if isPipe() {
		return os.Getenv("GO_PROMPTER_USE_PIPE") == ""
	}
	return false
}

func isPipe() bool {
	return !(isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())) ||
		!(isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd()))

}

func (p *Prompter) msg() string {
	msg := p.Message
	if len(p.Choices) > 0 {
		msg += fmt.Sprintf(" (%s)", strings.Join(p.Choices, "/"))
	}
	if p.Default != "" {
		msg += " [" + p.Default + "]"
	}
	return msg + ": "
}

func (p *Prompter) errorMsg() string {
	if p.Regexp != nil {
		return fmt.Sprintf("# Answer should match /%s/", p.Regexp)
	}
	if len(p.Choices) > 0 {
		if len(p.Choices) == 1 {
			return fmt.Sprintf("# Enter `%s`", p.Choices[0])
		}
		choices := make([]string, len(p.Choices)-1)
		for i, v := range p.Choices[:len(p.Choices)-1] {
			choices[i] = "`" + v + "`"
		}
		return fmt.Sprintf("# Enter %s or `%s`", strings.Join(choices, ", "), p.Choices[len(p.Choices)-1])
	}
	return ""
}

func (p *Prompter) inputIsValid(input string) bool {
	return p.regexp().MatchString(input)
}

var allReg = regexp.MustCompile(`.*`)

func (p *Prompter) regexp() *regexp.Regexp {
	if p.Regexp != nil {
		return p.Regexp
	}
	if p.reg != nil {
		return p.reg
	}
	if len(p.Choices) == 0 {
		p.reg = allReg
		return p.reg
	}

	choices := make([]string, len(p.Choices))
	for i, v := range p.Choices {
		choices[i] = regexp.QuoteMeta(v)
	}
	ignoreReg := ""
	if p.IgnoreCase {
		ignoreReg = "(?i)"
	}
	p.reg = regexp.MustCompile(fmt.Sprintf(`%s\A(?:%s)\z`, ignoreReg, strings.Join(choices, "|")))
	return p.reg
}
