FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/registry ./cmd/registry

FROM scratch
COPY --from=build /out/registry /registry
EXPOSE 8080
ENTRYPOINT ["/registry"]
