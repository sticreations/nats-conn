FROM golang:latest as build
ENV GO111MODULE=on
WORKDIR /natsconnector
COPY . . 
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-w -s" -o /usr/bin/producer


FROM alpine:3.9 as ship
RUN apk add --no-cache ca-certificates
COPY --from=build /usr/bin/producer /usr/bin/producer
WORKDIR /root/

CMD ["/usr/bin/producer"]
