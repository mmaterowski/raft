FROM golang:latest
LABEL maintainer="Michal Materowski <michal.materowski@gmail.com>"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o main .

EXPOSE 6969
EXPOSE 6960

CMD ["./main"]