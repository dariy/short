FROM cgr.dev/chainguard/go:latest AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /short .

FROM scratch
COPY --from=builder /short /short
COPY config/links.yml /config/links.yml
EXPOSE 8080
ENV CONFIG_PATH=/config/links.yml
ENV LISTEN_ADDR=:8080
ENTRYPOINT ["/short"]
