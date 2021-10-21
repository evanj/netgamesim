FROM golang:1.17.2-bullseye AS builder
COPY cloudrunhost.go /go/src/cloudrunhost/
WORKDIR /go/src/cloudrunhost
RUN go mod init github.com/evanj/netgamesim/cloudrunhost && go build --mod=readonly cloudrunhost.go && find /go/src -type f

FROM gcr.io/distroless/base-debian11:latest AS run
COPY --from=builder /go/src/cloudrunhost/cloudrunhost /cloudrunhost
COPY build/* /data/

# Use a non-root user: slightly more secure (defense in depth)
USER nobody
WORKDIR /
EXPOSE 8080
ENTRYPOINT ["/cloudrunhost"]
