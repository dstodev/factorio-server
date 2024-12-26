# Expects build context set to repo root

FROM python:3

COPY manage /repo/manage/

WORKDIR /repo/manage

RUN python -m venv /opt/venv \
	&& . /opt/venv/bin/activate \
	&& pip install --upgrade pip \
	&& pip install --no-cache-dir .

ENV PATH="/opt/venv/bin:$PATH"

ENTRYPOINT ["manage"]
