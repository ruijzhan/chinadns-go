FROM golang as builder
COPY . /build
WORKDIR /build
RUN make

FROM alpine:3.12.0
COPY --from=builder /build/chinadns /usr/local/bin/
WORKDIR /
