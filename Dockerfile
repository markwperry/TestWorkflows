FROM golang:1.25-alpine AS builder

ARG BUILD_VERSION=dev
ARG BUILD_GITHASH=unknown
ARG BRANCH=unknown

ENV BUILD_VERSION=${BUILD_VERSION}
ENV BUILD_GITHASH=${BUILD_GITHASH}
ENV BRANCH=${BRANCH}

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN chmod +x build.sh && ./build.sh

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
