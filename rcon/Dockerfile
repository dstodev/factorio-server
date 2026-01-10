FROM ubuntu:noble
#    ubuntu:latest tracks the latest LTS release

RUN apt-get update \
	&& apt-get install -y \
	build-essential \
	vim \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /root
COPY . .
RUN make test \
	&& make install \
	&& make clean

#ARG rcon_port=34240
#EXPOSE $rcon_port/tcp
