FROM golang:1.13 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
ARG GOPROXY
ARG GOSUMDB
RUN go mod download

COPY cmd cmd/
COPY internal internal/
COPY public public/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o q3 ./cmd/q3

FROM alpine:3.12 as quake-n-bake

RUN apk add --no-cache git gcc make libc-dev sdl2-dev mesa-dev
RUN git clone https://github.com/ioquake/ioq3
RUN cd /ioq3 && make
RUN cp /ioq3/build/release-linux-x86_64/ioq3ded.x86_64 /usr/local/bin/ioq3ded

FROM alpine:3.12

COPY --from=builder /workspace/q3 /usr/local/bin
COPY --from=quake-n-bake /usr/local/bin/ioq3ded /usr/local/bin
COPY --from=quake-n-bake /lib/ld-musl-x86_64.so.1 /lib

ENTRYPOINT ["/usr/local/bin/q3"]
