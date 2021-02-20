FROM golang:1.16

ENV GO111MODULE=on

# sqitch
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    sqitch libdbd-pg-perl postgresql-client \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /project

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o /bin/scribly ./cmd/site