package prompter

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-isatty"
	"golang.org/x/crypto/ssh/terminal"
)

// Prompter is object for prompting
type Prompter struct {
	Message    string
	Choices    []string
	IgnoreCase bool
	Default    string
	Regexp     *regexp.Regexp
	NoEcho     bool
	UseDefault bool
}

// Prompt displays a prompt and returns answer
func (p *Prompter) Prompt() string {
	if p.UseDefault || skip() {
		return p.Default
	}

	input := ""
	for {
		fmt.Print(p.msg())

		if p.NoEcho {
			b, err := terminal.ReadPassword(int(os.Stdin.Fd()))
			if err == nil {
				input = string(b)
			}
			fmt.Print("\n")
		} else {
			scanner := bufio.NewScanner(os.Stdin)
			ok := scanner.Scan()
			if ok {
				input = strings.TrimRight(scanner.Text(), "\r\n")
			}
		}
		if input == "" {
			input = p.Default
		}
		if p.inputIsValid(input) {
			break
		}
		fmt.Println(p.errorMsg())
	}
	return input
}

func skip() bool {
	if os.Getenv("GO_PROMPTER_USE_DEFAULT") != "" {
		return true
	}
	return !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd())
}

func (p *Prompter) msg() string {
	msg := p.Message
	if p.Choices != nil && len(p.Choices) > 0 {
		msg += fmt.Sprintf(" (%s)", strings.Join(p.Choices, "/"))
	}
	if p.Default != "" {
		msg += " [" + p.Default + "]"
	}
	return msg + ": "
}

func (p *Prompter) errorMsg() string {
	if p.Choices != nil && len(p.Choices) > 0 {
		if len(p.Choices) == 1 {
			return fmt.Sprintf("# Enter `%s`", p.Choices[0])
		}
		choices := make([]string, len(p.Choices)-1)
		for i, v := range p.Choices[:len(p.Choices)-1] {
			choices[i] = "`" + v + "`"
		}
		return fmt.Sprintf("# Enter %s or `%s`", strings.Join(choices, ", "), p.Choices[len(p.Choices)-1])
	}
	return fmt.Sprintf("# Answer should be matched the regexp: /%s/", p.regexp())
}

func (p *Prompter) inputIsValid(input string) bool {
	if p.IgnoreCase {
		input = strings.ToLower(input)
	}
	return p.regexp().MatchString(input)
}

var allReg = regexp.MustCompile(`.*`)

func (p *Prompter) regexp() *regexp.Regexp {
	if p.Regexp != nil {
		return p.Regexp
	}
	if p.Choices == nil || len(p.Choices) == 0 {
		p.Regexp = allReg
		return p.Regexp
	}
	choices := make([]string, len(p.Choices))

	for i, v := range p.Choices {
		choice := regexp.QuoteMeta(v)
		if p.IgnoreCase {
			choice = strings.ToLower(choice)
		}
		choices[i] = choice
	}
	regStr := fmt.Sprintf(`\A(?:%s)\z`, strings.Join(choices, "|"))
	p.Regexp = regexp.MustCompile(regStr)
	return p.Regexp
}
