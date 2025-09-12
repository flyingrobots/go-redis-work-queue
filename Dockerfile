# syntax=docker/dockerfile:1

FROM golang:1.21 AS build
WORKDIR /src
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/job-queue-system ./cmd/job-queue-system

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /
COPY --from=build /out/job-queue-system /job-queue-system
USER nonroot:nonroot
ENTRYPOINT ["/job-queue-system"]

