# syntax=docker/dockerfile:1
FROM rcon:latest

# set DEBIAN_FRONTEND as an ARG so it only affects the image build environment
ARG DEBIAN_FRONTEND=noninteractive

ENV PATH="$PATH:/usr/games"

# locale
RUN apt-get update \
	&& apt-get install --yes locales \
	&& locale-gen en_US.UTF-8 \
	&& update-locale LANG=en_US.UTF-8

# assorted tools & dependencies
RUN apt-get update \
	&& apt-get install --yes --install-recommends \
	ca-certificates \
	cabextract \
	curl \
	dbus \
	dbus-x11 \
	dconf-service \
	file \
	gnupg \
	gnutls-bin \
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
	p7zip \
	software-properties-common \
	tree \
	unrar \
	unzip \
	vim \
	websockify \
	wget \
	winbind \
	x11vnc \
	xdg-utils \
	xvfb \
	zenity

# Install steamcmd
RUN echo steam steam/license note '' | debconf-set-selections \
	&& echo steam steam/question select 'I AGREE' | debconf-set-selections \
	&& dpkg --add-architecture i386 \
	&& apt-add-repository multiverse \
	&& apt-get update \
	&& apt-get install --yes --install-recommends steamcmd \
	&& steamcmd +quit

# Install latest Wine
# https://gitlab.winehq.org/wine/wine/-/wikis/Debian-Ubuntu#noble
RUN mkdir -pm755 /etc/apt/keyrings \
	&& wget -O - https://dl.winehq.org/wine-builds/winehq.key \
	| gpg --dearmor -o /etc/apt/keyrings/winehq-archive.key - \
	&& dpkg --add-architecture i386 \
	&& wget -NP /etc/apt/sources.list.d/ \
	https://dl.winehq.org/wine-builds/ubuntu/dists/noble/winehq-noble.sources \
	&& apt-get update \
	&& apt-get install --yes --install-recommends \
	winehq-stable \
	wine-stable
# 	wine32:i386

# Prepare Gecko & Mono for Wine
# https://gitlab.winehq.org/wine/wine/-/wikis/Gecko
# https://gitlab.winehq.org/wine/wine/-/wikis/Wine-Mono
ARG GECKO_VERSION=2.47.4
ARG MONO_VERSION=10.4.1
RUN mkdir --parents /usr/share/wine/gecko \
	&& wget --directory-prefix=/usr/share/wine/gecko/ \
	https://dl.winehq.org/wine/wine-gecko/${GECKO_VERSION}/wine-gecko-${GECKO_VERSION}-x86.msi \
	https://dl.winehq.org/wine/wine-gecko/${GECKO_VERSION}/wine-gecko-${GECKO_VERSION}-x86_64.msi \
	&& mkdir --parents /usr/share/wine/mono \
	&& wget --directory-prefix=/usr/share/wine/mono/ \
	https://dl.winehq.org/wine/wine-mono/${MONO_VERSION}/wine-mono-${MONO_VERSION}-x86.msi

# Install latest Winetricks
# https://github.com/Winetricks/winetricks?tab=readme-ov-file#manual-install
RUN apt-get purge --yes winetricks \
	&& apt-get update \
	&& apt-get install --yes winetricks

ARG user_id
ARG user_name
ARG group_id
ARG group_name

RUN groupadd --gid "$group_id" "$group_name" \
	&& useradd --uid "$user_id" \
	--gid "$group_id" \
	--create-home \
	"$user_name"

ENV TZ=America/Los_Angeles
ENV DISPLAY=:99

ENV WINEARCH=wow64
ENV WINEDEBUG=-all
ENV WINEPREFIX=/opt/wineprefix

ARG WINETRICKS_LATEST_VERSION_CHECK=enabled
RUN winetricks --self-update

RUN mkdir --parents "$WINEPREFIX" \
	&& chown "$user_name:$group_name" "$WINEPREFIX"

# RUN apt-get clean \
# 	&& rm -rf /var/lib/apt/lists/*
RUN rm -rf /tmp/* /tmp/.??*

# https://stackoverflow.com/questions/40359282/launch-a-cat-command-unix-into-dockerfile
RUN <<-'EOF' cat >/usr/local/bin/tricks && chmod =rx,u+w /usr/local/bin/tricks
	#!/bin/bash
	set -e
	xvfb winetricks --unattended "$@" | sed '/saving to:/,/saved/{/saving to:/!{/saved/!d}}'
EOF

RUN <<-'EOF' cat >/usr/local/bin/xvfb && chmod =rx,u+w /usr/local/bin/xvfb
	#!/bin/bash
	set -e
	cleanup() {
	  find /tmp -user "$(id -u)" -delete
	}
	trap cleanup EXIT
	set -x
	xvfb-run --auto-servernum "$@"
EOF

USER "$user_name"
WORKDIR "/home/$user_name"

RUN xvfb wineboot --init

# WINETRICKS_CACHE does not seem to change where winetricks actually puts its
# files. Set to the directory winetricks uses & mount it below.
ARG WINETRICKS_CACHE="/home/$user_name/.cache/winetricks"
RUN mkdir --parents "$WINETRICKS_CACHE"

# https://docs.docker.com/reference/dockerfile/#run---mounttypecache
RUN --mount=type=cache,target="$WINETRICKS_CACHE",uid="$user_id",gid="$group_id" \
	tricks arch=64 win7

RUN --mount=type=cache,target="$WINETRICKS_CACHE",uid="$user_id",gid="$group_id" \
	tricks corefonts allfonts

# RUN --mount=type=cache,target="$WINETRICKS_CACHE",uid="$user_id",gid="$group_id" \
# 	ls -la "$WINETRICKS_CACHE" && false

# https://forum.winehq.org/viewtopic.php?t=35278
# RUN --mount=type=cache,target="$WINETRICKS_CACHE",uid="$user_id",gid="$group_id" \
# 	tricks comctl32ocx comdlg32ocx ole32 oleaut32
# # 	comctl32

RUN --mount=type=cache,target="$WINETRICKS_CACHE",uid="$user_id",gid="$group_id" \
	tricks vcrun2015 mfc140 dotnet48 directx9

# RUN --mount=type=cache,target="$WINETRICKS_CACHE",uid="$user_id",gid="$group_id" \
# 	tricks d3dx9 d3dx10 d3dcompiler_47 d3dx11_43

# RUN --mount=type=cache,target="$WINETRICKS_CACHE",uid="$user_id",gid="$group_id" \
# 	tricks msxml4 msxml6

# RUN --mount=type=cache,target="$WINETRICKS_CACHE",uid="$user_id",gid="$group_id" \
# 	tricks xact xact_x64 xinput

# RUN --mount=type=cache,target="$WINETRICKS_CACHE",uid="$user_id",gid="$group_id" \
# 	tricks gdiplus riched20

# comctl32 \
# arch=64 \
# win7 \

# arch=64 \
# win10 \
# vcrun2015 \
# allfonts \
# d3dx9 \
# d3dx10 \
# d3dcompiler_47 \
# d3dx11_43 \
# dotnet48 \
# ole32 \
# oleaut32

# d3dx9 dxsdk_jun2010
# vcrun2015 vcrun2017 vcrun2019 vcrun2022

# https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist?view=msvc-170#latest-supported-redistributable-version
# RUN wget https://aka.ms/vc14/vc_redist.x64.exe -O /tmp/vc14_redist.exe \
# 	&& xvfb-run wine /tmp/vc14_redist.exe /install /passive /norestart \
# 	&& rm /tmp/vc14_redist.exe

ARG game_port
EXPOSE $game_port/tcp
EXPOSE $game_port/udp
