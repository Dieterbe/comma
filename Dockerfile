FROM golang:1.26-alpine AS build
WORKDIR /src

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/* .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w" -o /out/comma .

FROM alpine:3.24
RUN mkdir /data && adduser -D -u 10001 app && chown app:app /data

COPY --from=build /out/comma /usr/local/bin/comma

USER app

EXPOSE 9991

ENTRYPOINT ["comma", "/data", "9991"]