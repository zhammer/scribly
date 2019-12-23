FROM python:3.7-alpine

COPY . ./project
WORKDIR /project

RUN apk add --no-cache --virtual .build-deps \
    gcc musl-dev libffi-dev libc-dev make libxslt-dev \
    && pip install --no-cache-dir -r requirements.txt \
    && apk del .build-deps
