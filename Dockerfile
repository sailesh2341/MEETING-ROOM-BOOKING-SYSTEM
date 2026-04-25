FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /meeting-room-booking ./cmd/launcher
RUN go build -o /booking-cli ./cmd/booking-cli

FROM alpine:3.20

WORKDIR /app

COPY --from=build /meeting-room-booking /usr/local/bin/meeting-room-booking
COPY --from=build /booking-cli /usr/local/bin/booking-cli

EXPOSE 8080

CMD ["meeting-room-booking"]
