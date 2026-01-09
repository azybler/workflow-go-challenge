FROM golang:1.25-alpine

WORKDIR /app

ENV CGO_ENABLED=0

COPY go.mod go.sum ./

RUN go mod download

# using air for hot reload.
RUN go install github.com/air-verse/air@v1.63.5

# Copy source code
COPY . .

# Expose port 8080
EXPOSE 8080

# Command to run the application with hot reloading
CMD ["air", "-c", ".air.toml"]
