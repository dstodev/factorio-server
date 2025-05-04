FROM rcon:latest

RUN apt update \
	&& DEBIAN_FRONTEND=noninteractive \
	# apt install -y \
	# 	wget \
	&& apt clean

ARG game_port
EXPOSE $game_port/udp

WORKDIR /game/hot/
#       /game/hot/ is mounted as a bind volume
