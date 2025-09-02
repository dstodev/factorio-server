CLIENT := rcon
SRC := main.cxx

PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin

CXX := g++
CXXFLAGS := -std=c++17 -O3 -Wall -Wextra

DOCKERFILE := Dockerfile
IMAGE_TAG := rcon:latest

all: image
.PHONY: all

$(CLIENT): $(SRC)
	$(CXX) $(CXXFLAGS) -o $@ $^

test: $(CLIENT)
	./$< test
.PHONY: test

install: $(CLIENT)
	install -Dm755 -t $(BINDIR) $<
.PHONY: install

clean:
	rm -f $(CLIENT)
.PHONY: clean

image:
	docker build -f $(DOCKERFILE) -t $(IMAGE_TAG) .
.PHONY: image

shell: image
	docker run -it --rm $(IMAGE_TAG)
.PHONY: shell
