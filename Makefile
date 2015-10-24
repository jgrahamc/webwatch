NAME       := webwatch

include gmsl/gmsl

PWD := $(shell pwd)

.PHONY: all
all: $(NAME)

.PHONY: $(NAME)
$(NAME): bin/$(NAME)

.PHONY: bin/$(NAME)
bin/$(NAME): ; @GOPATH="${PWD}" go install $(NAME)

.PHONY: clean
clean:
	GOPATH="${PWD}" go clean
