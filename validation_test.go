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
			tt.prompter.inputIsValid("a")
			tt.update(&tt.prompter)
			if !tt.prompter.inputIsValid(tt.accepted) {
				t.Errorf("inputIsValid(%q) = false, want true", tt.accepted)
			}
			if tt.rejected != "" && tt.prompter.inputIsValid(tt.rejected) {
				t.Errorf("inputIsValid(%q) = true, want false", tt.rejected)
			}
		})
	}
}

func TestRegexpOverrideUsesCurrentChoices(t *testing.T) {
	p := &Prompter{Choices: []string{"a"}}
	p.inputIsValid("a")
	p.Regexp = regexp.MustCompile(`\Acustom\z`)
	p.Choices = []string{"b"}
	if !p.inputIsValid("custom") || p.inputIsValid("b") {
		t.Fatal("Regexp should take priority over Choices")
	}
	p.Regexp = nil
	if !p.inputIsValid("b") || p.inputIsValid("a") || p.inputIsValid("custom") {
		t.Fatal("validation should use current Choices after removing Regexp")
	}
}
