FROM ubuntu:latest
#    ubuntu:latest tracks the latest LTS release

RUN apt-get update \
	&& apt-get install -y \
	build-essential \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /root
COPY . .
RUN make test \
	&& make install \
	&& make clean

#ARG rcon_port=34240
#EXPOSE $rcon_port/tcp

ARG user_id=30120
ARG user_name=rcon
ARG group_id=30120
ARG group_name=rcon

RUN groupadd --gid $group_id $group_name \
	&& useradd --uid $user_id \
	--gid $group_id \
	--create-home \
	$user_name

USER $user_name
WORKDIR /home/$user_name
