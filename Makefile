GO = go
GOLINT = golint

.PHONY: all test clean lint

all: test

clean:
	$(GO) clean ./

lint:
	$(GOLINT) ./

test:
	$(GO) test -v ./
