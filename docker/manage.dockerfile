FROM python:3

WORKDIR /repo/

ENTRYPOINT [ "docker/manage.entrypoint.sh" ]

ARG user_id
ARG user_name
ARG group_id
ARG group_name

RUN groupadd --gid $group_id $group_name \
	&& useradd --create-home \
	           --uid $user_id \
	           --gid $group_id \
	           $user_name
