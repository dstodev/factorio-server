FROM rcon:latest

USER root

RUN apt update \
	&& DEBIAN_FRONTEND=noninteractive \
	apt install -y \
	openjdk-21-jre-headless \
	inotify-tools \
	wget \
	unzip \
	&& apt clean

ARG user_name
USER $user_name

ARG game_port
EXPOSE $game_port/udp
