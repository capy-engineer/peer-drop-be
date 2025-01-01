FROM golang:1.23-alpine AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd
# -a -installsuffix cgo

FROM alpine
WORKDIR /
COPY --from=build /app .
CMD ["/app"]