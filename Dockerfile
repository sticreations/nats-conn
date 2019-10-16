FROM golang:1.12.6-alpine as build
RUN apk add git
WORKDIR /natsconnector
COPY ./config .
COPY ./nats .
COPY go.mod . 
COPY go.sum . 
COPY main.go .
RUN go get -d -v ./...
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-w -s" -o /usr/bin/producer


FROM alpine:3.9 as ship
RUN apk add --no-cache ca-certificates
COPY --from=build /usr/bin/producer /usr/bin/producer
WORKDIR /root/

CMD ["/usr/bin/producer"]
