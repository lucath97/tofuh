FROM golang:1.24-alpine3.21 AS build
WORKDIR /build
COPY . .
RUN go build -o ./broker cmd/app/tofuh.go

FROM alpine:3.21
WORKDIR /app
COPY --from=build /build/broker . 
CMD ["/app/broker"]