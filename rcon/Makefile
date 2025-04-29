CLIENT := client.out
SRC := main.cxx

CXX := g++
CXXFLAGS := -std=c++17 -O3 -Wall -Wextra

DOCKERFILE := rcon.dockerfile
DOCKER_IMAGE := rcon:latest

all: $(CLIENT) image

$(CLIENT): $(SRC)
	$(CXX) $(CXXFLAGS) -o $@ $^

test: $(CLIENT)
	./$< test

clean:
	rm -f $(CLIENT)

image:
	docker build -f $(DOCKERFILE) -t $(DOCKER_IMAGE) .

shell: image
	docker run -it --rm $(DOCKER_IMAGE)

.PHONY: all test clean image shell
