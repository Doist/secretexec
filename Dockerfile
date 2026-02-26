FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:alpine AS builder
WORKDIR /app
ENV CGO_ENABLED=0 GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETARCH
RUN GOARCH=$TARGETARCH go build -ldflags="-w -s" -trimpath

FROM scratch
COPY --from=builder /app/secretexec /var/runtime/
VOLUME ["/var/runtime"]
CMD ["/var/runtime/secretexec", "-h"]
