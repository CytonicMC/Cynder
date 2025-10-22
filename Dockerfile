FROM --platform=$BUILDPLATFORM golang:1.25.3 AS build

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY cynder ./cynder
COPY gate.go ./

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-s -w" -a -o Cynder gate.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=build /workspace/Cynder /
COPY config.yml /
CMD ["/Cynder"]