FROM rcon:latest

RUN apt update \
	&& DEBIAN_FRONTEND=noninteractive \
	apt install -y \
	openjdk-17-jre-headless \
	inotify-tools \
	wget \
	unzip \
	&& apt clean

ARG user_id
ARG user_name
ARG group_id
ARG group_name

RUN groupadd --gid $group_id $group_name \
	&& useradd --uid $user_id \
	--gid $group_id \
	--create-home \
	$user_name

USER $user_name
WORKDIR /home/$user_name

ARG game_port
EXPOSE $game_port/udp
