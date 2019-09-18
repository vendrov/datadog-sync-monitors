# Build the manager binary
FROM golang:1.12.5 as builder

WORKDIR /workspace

# Copy the go source
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o start main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:latest
WORKDIR /
COPY --from=builder /workspace/start .
# ENTRYPOINT ["/start"]
