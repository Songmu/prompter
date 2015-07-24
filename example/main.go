package main

import (
	"fmt"
	"regexp"

	"github.com/Songmu/prompter"
)

func main() {
	input := (&prompter.Prompter{
		Choices:    []string{"aa", "bb", "cc"},
		Default:    "aa",
		Message:    "plaase select",
		IgnoreCase: true,
	}).Prompt()
	fmt.Println("your input is " + input)

	input = (&prompter.Prompter{
		Message: "enter password",
		Regexp:  regexp.MustCompile(`.{8,}`),
		NoEcho:  true,
	}).Prompt()
	fmt.Println("your password is " + input)
}
