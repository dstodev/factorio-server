CLIENT := rcon
SRC := main.cxx

PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin

CXX := g++
CXXFLAGS := -std=c++17 -O3 -Wall -Wextra

DOCKERFILE := Dockerfile
DOCKER_IMAGE := rcon:latest

all: docker

$(CLIENT): $(SRC)
	$(CXX) $(CXXFLAGS) -o $@ $^

test: $(CLIENT)
	./$< test

install: $(CLIENT)
	install -Dm755 -t $(BINDIR) $<

clean:
	rm -f $(CLIENT)

docker: image alpine ubuntu

image: ubuntu
	docker build -f $(DOCKERFILE) -t $(DOCKER_IMAGE) .

alpine:
	docker pull alpine:latest

ubuntu:
	docker pull ubuntu:latest

shell: image
	docker run -it --rm $(DOCKER_IMAGE)

.PHONY: all test install clean \
        docker image alpine ubuntu shell
