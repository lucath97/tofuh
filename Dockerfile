FROM golang:1.24-alpine3.21 AS build
WORKDIR /build
COPY . .
RUN go build -o ./msgcontrol ./cmd/msgcontrol/main.go
RUN go build -o ./httpcontrol ./cmd/httpcontrol/main.go  

FROM alpine:3.21 AS msgcontrol
WORKDIR /app
COPY --from=build /build/msgcontrol /app/ 
CMD ["/app/msgcontrol"]

FROM alpine:3.21 AS httpcontrol
WORKDIR /app
COPY --from=build /build/httpcontrol /app/
CMD [ "/app/httpcontrol" ]