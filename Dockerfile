FROM golang:1.19 as build-env

ADD . /go/src/github.com/a1comms/gcs-tftp-server
WORKDIR /go/src/github.com/a1comms/gcs-tftp-server

ARG CGO_ENABLED=0

RUN go mod vendor
RUN go build -ldflags "-s -w" -o /go/bin/app

FROM gcr.io/distroless/static-debian11
COPY --from=build-env /go/bin/app /
CMD ["/app"]
