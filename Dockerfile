FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN make build-for-docker

FROM alpine:3.20

COPY --from=builder /streams-server /streams-server

ENTRYPOINT ["/streams-server"]
