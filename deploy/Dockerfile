FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /omni-5e ./cmd/omni-5e

FROM gcr.io/distroless/static-debian12

COPY --from=builder /omni-5e /omni-5e

EXPOSE 8080

ENTRYPOINT ["/omni-5e"]
CMD ["serve"]
