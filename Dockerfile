FROM golang:alpine AS builder

ARG APP
WORKDIR /build
COPY . .
ENV CGO_ENABLED=0
RUN go build -a -tags netgo -ldflags '-w' -o app github.com/pav5000/reverse-redirector/cmd/${APP}

FROM scratch

COPY --from=builder /build/app /app
ENTRYPOINT ["/app"]
