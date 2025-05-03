CLIENT := rcon
SRC := main.cxx

CXX := g++
CXXFLAGS := -std=c++17 -O3 -Wall -Wextra

DOCKERFILE := Dockerfile
DOCKER_IMAGE := rcon:latest

all: image

$(CLIENT): $(SRC)
	$(CXX) $(CXXFLAGS) -o $@ $^

test: $(CLIENT)
	./$< test

clean:
	rm -f $(CLIENT)

install: $(CLIENT)
	install -Dm755 -t /usr/local/bin $<

image:
	docker build -f $(DOCKERFILE) -t $(DOCKER_IMAGE) .

shell: image
	docker run -it --rm $(DOCKER_IMAGE)

.PHONY: all test clean image shell
