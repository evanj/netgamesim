FROM golang:1.20.1-bullseye AS builder
COPY cloudrunhost.go /go/src/cloudrunhost/
WORKDIR /go/src/cloudrunhost
RUN go mod init github.com/evanj/netgamesim/cloudrunhost && go build cloudrunhost.go

FROM gcr.io/distroless/base-debian11:nonroot AS run
COPY --from=builder /go/src/cloudrunhost/cloudrunhost /cloudrunhost
COPY build/* /data/

# Use a non-root user: slightly more secure (defense in depth)
USER nobody
WORKDIR /
EXPOSE 8080
ENTRYPOINT ["/cloudrunhost"]
