prompter
=======

[![Test Status](https://github.com/Songmu/prompter/workflows/test/badge.svg?branch=main)][actions]
[![Coverage Status](https://codecov.io/gh/Songmu/prompter/branch/main/graph/badge.svg)][codecov]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![GoDev](https://pkg.go.dev/badge/github.com/Songmu/prompter)][godev]

[actions]: https://github.com/Songmu/prompter/actions?workflow=test
[codecov]: https://codecov.io/gh/Songmu/prompter
[license]: https://github.com/Songmu/prompter/blob/main/LICENSE
[godev]: http://pkg.go.dev/github.com/Songmu/prompter

## Description

utility for easy prompting in Golang

## Synopsis

```go
twitterID := prompter.Prompt("Enter your twitter ID", "")
lang := prompter.Choose("Which language do you like the most?", []string{"Perl", "Golang", "Scala", "Ruby"}, "Perl")
passwd := prompter.Password("Enter your password")
var likeSushi bool = prompter.YN("Do you like sushi?", true)
var likeBeer bool = prompter.YesNo("Do you like beer?", false)
```

## Features

- Easy to use
- Prompts and validation messages go to stderr, keeping stdout clean
- Care non-interactive (not a tty) environment
  - `Default` is used and the process is not blocked
- No howeyc/gopass (which uses cgo) dependency
  - cross build friendly
- Customizable prompt setting by using `&prompter.Prompter{}` directly

## Read errors

Use `Prompter.PromptWithError()` to distinguish an answer from EOF or a read
error:

```go
answer, err := (&prompter.Prompter{
    Message: "Enter your name",
    Default: "guest",
}).PromptWithError()
if err != nil {
    return err
}
fmt.Println(answer)
```

An empty line uses `Default`, but EOF and read errors return an empty answer and
an error. `Prompt()` and the convenience functions instead report read errors to
stderr and return an empty answer (`false` for `YN` and `YesNo`).
Explicit default mode and non-interactive mode still return `Default` without
reading input.

## License

[MIT][license]

## Author

[Songmu](https://github.com/Songmu)
