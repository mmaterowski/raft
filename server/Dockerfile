## Build
FROM golang:1.16-buster AS build
LABEL maintainer="Michal Materowski <michal.materowski@gmail.com>"
WORKDIR /app
 
# Download dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

RUN go install github.com/mattn/go-sqlite3

COPY . .

RUN go build -o /raft 

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /raft /raft

EXPOSE 6969
EXPOSE 6960
ARG env="Local"

USER nonroot:nonroot

ENTRYPOINT ["/raft"]