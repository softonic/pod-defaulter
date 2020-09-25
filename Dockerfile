FROM golang:1.14-buster AS build

ENV GOBIN=$GOPATH/bin

ADD . /src/pod-defaulter

WORKDIR /src/pod-defaulter

RUN make build

FROM debian:buster-slim

COPY --from=build /src/pod-defaulter/pod-defaulter /pod-defaulter

ENTRYPOINT ["/pod-defaulter"]
