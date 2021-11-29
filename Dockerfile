FROM golang:1.17 as build
WORKDIR /project
COPY go.mod .
RUN go mod download
COPY . /project
RUN go build ./backendv2/cmd/main.go

FROM ubuntu:latest as api-server
RUN apt update && apt install ca-certificates -y && rm -rf /var/cache/apt/*
COPY --from=build /project/main /
CMD ["./main"]