FROM rcon:latest

# set DEBIAN_FRONTEND as an ARG so it only affects the image build environment
ARG DEBIAN_FRONTEND=noninteractive

ENV PATH="$PATH:/usr/games"

RUN apt-get update && apt-get install -y \
	software-properties-common \
	locales \
	&& locale-gen en_US.UTF-8 \
	&& update-locale LANG=en_US.UTF-8 \
	&& echo steam steam/license note '' | debconf-set-selections \
	&& echo steam steam/question select 'I AGREE' | debconf-set-selections \
	&& dpkg --add-architecture i386 \
	&& apt-add-repository multiverse \
	&& apt-get update \
	&& apt-get install -y steamcmd \
	&& steamcmd +quit

RUN apt-get update && apt-get install -y \
	inotify-tools \
	jq \
	vim \
	xvfb \
	x11vnc \
	wine

RUN apt-get clean

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
