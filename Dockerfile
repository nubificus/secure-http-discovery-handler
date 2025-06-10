FROM harbor.nbfc.io/proxy_cache/library/golang:1.24.2-alpine3.21 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && \
    go mod vendor && \
    go mod verify
    
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -ldflags "-extldflags '-static'" -o discovery-handler ./cmd/secure-http-discovery-handler
    
FROM harbor.nbfc.io/proxy_cache/library/alpine:3.21

COPY --from=builder /app/discovery-handler /discovery-handler
ENTRYPOINT ["/discovery-handler"]