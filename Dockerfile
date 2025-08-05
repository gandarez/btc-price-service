FROM golang:1.24-bullseye AS build

WORKDIR /src

# Copy everything but defined in docker ignore file
COPY . .

RUN go mod vendor
RUN make build-linux-amd64

#####################
# Build final image #
#####################
FROM alpine AS bin

COPY --from=build /src/build/btc-price-service-linux-amd64 ./btc-price-service
COPY --from=build /src/configs ./configs

ENTRYPOINT ["./btc-price-service"]
