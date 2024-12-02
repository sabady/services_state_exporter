# Use a lightweight image with Go
FROM golang:1.23-alpine as builder

WORKDIR /app
COPY . .

# Build the binary
RUN go build -o swarm_exporter .

# Final image
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/swarm_exporter .

EXPOSE 9180
ENTRYPOINT ["./swarm_exporter"]
