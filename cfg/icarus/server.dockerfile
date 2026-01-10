FROM rcon:latest

# set DEBIAN_FRONTEND as an ARG so it only affects the image build environment
ARG DEBIAN_FRONTEND=noninteractive

ENV PATH="$PATH:/usr/games"

# basic dependencies
RUN apt-get update && apt-get install --yes \
	ca-certificates \
	curl \
	tini \
	vim \
	wget

# install steamcmd
RUN apt-get install --yes \
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

# assorted dependencies
RUN apt-get install --yes \
	dbus \
	dbus-x11 \
	dconf-service \
	inotify-tools \
	jq \
	libappindicator3-1 \
	libasound2t64 \
	libgbm1 \
	libgtk-3-0 \
	libnotify4 \
	libnss3 \
	libsecret-1-0 \
	libxdamage1 \
	libxrandr2 \
	libxss1 \
	novnc \
	openbox \
	websockify \
	winbind \
	wine \
	winetricks \
	x11vnc \
	xdg-utils \
	xvfb

RUN apt-get clean && rm -rf /var/lib/apt/lists/*

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

ENV DISPLAY=:99

ENV WINEARCH=win64
ENV WINEPREFIX=/home/$user_name/.wine

RUN xvfb-run winetricks --unattended vcrun2019
