FROM python:3

RUN apt-get update \
	&& DEBIAN_FRONTEND=noninteractive \
	apt-get install -y --no-install-recommends \
	tree \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /repo/

ENTRYPOINT ["/bin/bash", "docker/py.entrypoint.sh"]

ARG user_id
ARG user_name
ARG group_id
ARG group_name

RUN groupadd --gid $group_id $group_name \
	&& useradd --create-home --uid $user_id --gid $group_id $user_name

USER $user_id:$group_id
