FROM golang:1.13 AS builder

USER root
RUN mkdir -p /app
WORKDIR /app

COPY . /app
RUN go get ./...
RUN go build ./cmd/kubeprober
RUN chmod +x kubeprober


# -------------------------------------------------
FROM centos:7

COPY --from=builder /app/kubeprober /kubeprober
ENTRYPOINT [ "/kubeprober" ]
