FROM golang:1.10 AS builder

WORKDIR /go/src/kata-pod-annotate

COPY . ./
RUN go get ./... && CGO_ENABLED=0 go build -o /go/bin/kata-pod-annotate

FROM alpine:3.7
COPY --from=builder /go/bin/kata-pod-annotate /kata-pod-annotate
ENTRYPOINT ["/kata-pod-annotate"]

