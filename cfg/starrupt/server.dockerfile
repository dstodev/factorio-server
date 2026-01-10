FROM rcon:latest

# set DEBIAN_FRONTEND as an ARG so it only affects the image build environment
ARG DEBIAN_FRONTEND=noninteractive

ENV PATH="$PATH:/usr/games"

# locale
RUN apt-get update && apt-get install --yes \
	locales \
	&& locale-gen en_US.UTF-8 \
	&& update-locale LANG=en_US.UTF-8

# assorted tools & dependencies
RUN apt-get update && apt-get install --yes \
	ca-certificates \
	curl \
	dbus \
	dbus-x11 \
	dconf-service \
	file \
	inotify-tools \
	jq \
	less \
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
	software-properties-common \
	tree \
	vim \
	websockify \
	wget \
	winbind \
	wine \
	winetricks \
	x11vnc \
	xdg-utils \
	xvfb

# 32-bit stuff
RUN dpkg --add-architecture i386 \
	&& apt-get update \
	&& apt-get install --yes \
	wine32:i386

# install steamcmd
RUN echo steam steam/license note '' | debconf-set-selections \
	&& echo steam steam/question select 'I AGREE' | debconf-set-selections \
	&& apt-add-repository multiverse \
	&& apt-get update \
	&& apt-get install -y steamcmd \
	&& steamcmd +quit

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

ENV TZ=America/Los_Angeles

ENV DISPLAY=:99

ENV WINEARCH=win64
ENV WINEPREFIX=/home/$user_name/.wine

RUN winetricks --self-update

USER $user_name
WORKDIR /home/$user_name

RUN winetricks arch=64
RUN xvfb-run winetricks --unattended --force win10 vcrun2022 corefonts

# vcrun2015 vcrun2017 vcrun2019 vcrun2022

# https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist?view=msvc-170#latest-supported-redistributable-version
# RUN wget https://aka.ms/vc14/vc_redist.x64.exe -O /tmp/vc14_redist.exe \
# 	&& xvfb-run wine /tmp/vc14_redist.exe /install /passive /norestart \
# 	&& rm /tmp/vc14_redist.exe

# USER root

ARG game_port
EXPOSE $game_port/tcp
EXPOSE $game_port/udp
