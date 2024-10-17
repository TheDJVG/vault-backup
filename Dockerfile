FROM golang:1.23-bookworm as builder
COPY . .
RUN CGO_ENABLED=0 go build -o /vault-backup ./cmd/vault-backup

FROM gcr.io/distroless/static-debian12
COPY --from=builder /vault-backup /
ENTRYPOINT ["/vault-backup"]
