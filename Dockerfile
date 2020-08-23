FROM golang as builder
COPY . /build
WORKDIR /build
RUN make && \
    make chnroute

FROM alpine:3.12.0
COPY --from=builder /build/chinadns /usr/local/bin/
COPY --from=builder /build/chnroutegen /usr/local/bin/
COPY --from=builder /build/chnroute.json /
WORKDIR /
#ENTRYPOINT ["/chinadns"]
