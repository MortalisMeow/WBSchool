FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /wb-school ./cmd/main.go

FROM alpine:latest

COPY --from=builder /wb-school /wb-school
COPY ./ui/html /ui/html
COPY ./ui/static /ui/static

WORKDIR /
EXPOSE 8081
CMD ["/wb-school"]