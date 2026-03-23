FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY utils ./utils
COPY configs ./configs

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/tasktracker-api ./cmd/api

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/tasktracker-api ./tasktracker-api

ENV SERVERPORT=8080
ENV SERVERIP=0.0.0.0

EXPOSE 8080

CMD ["./tasktracker-api"]
