FROM golang:1.20-alpine3.16 AS build

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN go build -o file-dance .

FROM alpine:3.16 AS runtime

COPY --from=build /app/file-dance ./file-dance

