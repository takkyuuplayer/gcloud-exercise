FROM golang:1.21-bookworm as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o server

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/server /server

CMD ["/server"]
