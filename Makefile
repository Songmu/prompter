ifdef update
  u=-u
endif

export GO111MODULE=on

.PHONY: deps
deps:
	go get ${u} -d
	go mod tidy

.PHONY: devel-deps
devel-deps: deps
	GO111MODULE=off go get ${u}  \
	  golang.org/x/lint/golint   \
	  github.com/mattn/goveralls \
	  github.com/Songmu/godzil/cmd/godzil

.PHONY: test
test: deps
	go test

.PHONY: lint
lint: devel-deps
	go vet
	golint -set_exit_status

.PHONY: cover
cover: devel-deps
	goveralls

.PHONY: devel-deps
release: devel-deps
	godzil release
