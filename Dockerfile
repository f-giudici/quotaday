FROM golang:1.24 AS builder

ENV CGO_ENABLED=0
WORKDIR /src
ADD go.mod .
ADD go.sum .
RUN go mod download
ADD main.go .
ADD Makefile .
ADD cmd cmd
ADD pkg pkg
COPY main.go /app/main.go

ARG VERSION=v0.0.0
ARG COMMIT=""
RUN go build \
    -ldflags "-w -s \
    -X github.com/f-giudici/quotaday/cmd.version=${VERSION} \
    -X github.com/f-giudici/quotaday/cmd.gitCommit=${COMMIT}" \
    -o /quotaday


FROM scratch
WORKDIR /
COPY --from=builder /quotaday .

ENTRYPOINT [ "/quotaday" ]
