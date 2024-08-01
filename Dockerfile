FROM golang:1.22.0-alpine3.19 AS build

WORKDIR /src
COPY . /src/
RUN go build -ldflags="-s -w"

FROM alpine:3.19
COPY --from=build /src/route-sync /route-sync
ENTRYPOINT [ "/route-sync" ]
