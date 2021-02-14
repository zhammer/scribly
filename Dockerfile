FROM python:3.9

COPY . ./project
WORKDIR /project

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    sqitch libdbd-pg-perl postgresql-client \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir -r requirements.txt
