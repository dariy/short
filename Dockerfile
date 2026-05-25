FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /short .

FROM scratch
COPY --from=builder /short /short
EXPOSE 8080
ENV CONFIG_PATH=/config/links.yml
ENV LISTEN_ADDR=:8080
ENTRYPOINT ["/short"]
