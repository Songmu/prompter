package prompter

import (
	"regexp"
	"testing"
)

func TestValidationUsesCurrentSettings(t *testing.T) {
	tests := []struct {
		name     string
		prompter Prompter
		update   func(*Prompter)
		accepted string
		rejected string
	}{
		{
			name:     "replace choices",
			prompter: Prompter{Choices: []string{"a"}},
			update:   func(p *Prompter) { p.Choices = []string{"b"} },
			accepted: "b",
			rejected: "a",
		},
		{
			name:     "mutate choices",
			prompter: Prompter{Choices: []string{"a"}},
			update:   func(p *Prompter) { p.Choices[0] = "b" },
			accepted: "b",
			rejected: "a",
		},
		{
			name:     "add choices",
			update:   func(p *Prompter) { p.Choices = []string{"b"} },
			accepted: "b",
			rejected: "a",
		},
		{
			name:     "remove choices",
			prompter: Prompter{Choices: []string{"a"}},
			update:   func(p *Prompter) { p.Choices = nil },
			accepted: "anything",
		},
		{
			name:     "enable ignore case",
			prompter: Prompter{Choices: []string{"a"}},
			update:   func(p *Prompter) { p.IgnoreCase = true },
			accepted: "A",
			rejected: "b",
		},
		{
			name:     "disable ignore case",
			prompter: Prompter{Choices: []string{"a"}, IgnoreCase: true},
			update:   func(p *Prompter) { p.IgnoreCase = false },
			accepted: "a",
			rejected: "A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prompter.clone().inputIsValid("a")
			tt.update(&tt.prompter)
			prompt := tt.prompter.clone()
			if !prompt.inputIsValid(tt.accepted) {
				t.Errorf("inputIsValid(%q) = false, want true", tt.accepted)
			}
			if tt.rejected != "" && prompt.inputIsValid(tt.rejected) {
				t.Errorf("inputIsValid(%q) = true, want false", tt.rejected)
			}
		})
	}
}

func TestRegexpOverrideUsesCurrentChoices(t *testing.T) {
	p := &Prompter{Choices: []string{"a"}}
	p.clone().inputIsValid("a")
	p.Regexp = regexp.MustCompile(`\Acustom\z`)
	p.Choices = []string{"b"}
	prompt := p.clone()
	if !prompt.inputIsValid("custom") || prompt.inputIsValid("b") {
		t.Fatal("Regexp should take priority over Choices")
	}
	p.Regexp = nil
	prompt = p.clone()
	if !prompt.inputIsValid("b") || prompt.inputIsValid("a") || prompt.inputIsValid("custom") {
		t.Fatal("validation should use current Choices after removing Regexp")
	}
}

func TestPromptCloneIsolatesSettings(t *testing.T) {
	base := &Prompter{
		Message:    "Choice",
		Choices:    []string{"a"},
		IgnoreCase: true,
		Default:    "a",
	}
	base.regexp()
	prompt := base.clone()
	if prompt == base {
		t.Fatal("clone must return a separate Prompter")
	}
	if prompt.reg != nil {
		t.Fatal("clone must not inherit the base's cached regexp")
	}
	base.Message = "Changed"
	base.Choices[0] = "b"
	base.IgnoreCase = false
	base.Default = "b"
	if got := prompt.msg(); got != "Choice (a) [a]: " {
		t.Errorf("clone message = %q, want original settings", got)
	}
	if !prompt.inputIsValid("A") || prompt.inputIsValid("b") {
		t.Fatal("clone must validate against its own settings")
	}
	if base.Choices[0] != "b" || base.IgnoreCase || base.Default != "b" {
		t.Fatal("validating a clone must not change the base")
	}
}

func TestPromptCloneReusesRegexp(t *testing.T) {
	base := &Prompter{Choices: []string{"a"}}
	prompt := base.clone()
	reg := prompt.regexp()
	if reg == nil || prompt.reg != reg {
		t.Fatal("clone must cache its generated regexp")
	}
	if !prompt.inputIsValid("a") || prompt.inputIsValid("b") {
		t.Fatal("clone must validate against its choices")
	}
	if prompt.regexp() != reg {
		t.Fatal("clone must reuse its regexp across validation attempts")
	}
	if base.reg != nil {
		t.Fatal("validating a clone must not populate the base's cache")
	}
	next := base.clone()
	if next.reg != nil || next.regexp() == reg {
		t.Fatal("a new clone must build its own regexp")
	}
}

func TestPromptUsesFreshClone(t *testing.T) {
	p := &Prompter{Message: "Choice", Choices: []string{"a"}}
	for _, tt := range []struct {
		choice    string
		withError bool
	}{
		{choice: "a"},
		{choice: "b", withError: true},
	} {
		t.Run(tt.choice, func(t *testing.T) {
			promptStreams(t, tt.choice+"\n")
			p.Choices[0] = tt.choice
			var got string
			if tt.withError {
				var err error
				got, err = p.PromptWithError()
				if err != nil {
					t.Fatal(err)
				}
			} else {
				got = p.Prompt()
			}
			if got != tt.choice {
				t.Errorf("prompt answer = %q, want %q", got, tt.choice)
			}
			if p.reg != nil {
				t.Fatal("prompting must leave the base's cache empty")
			}
		})
	}
}
