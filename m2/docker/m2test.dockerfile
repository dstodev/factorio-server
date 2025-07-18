FROM alpine:latest

ENV USER=testuser
ENV GROUP=testgroup
ENV UID=30120
ENV GID=$UID

RUN addgroup \
	--gid "$GID" \
	"$GROUP" \
	&& adduser \
	--uid "$UID" \
	--ingroup "$GROUP" \
	--disabled-password \
	--gecos "" \
	"$USER"

RUN apk add --no-cache \
	coreutils \
	util-linux

USER $USER
