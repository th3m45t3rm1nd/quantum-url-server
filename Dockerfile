FROM golang:1.27rc2-alpine3.23
WORKDIR /usr/local/server

# Install the application dependencies
COPY go.mod go.sum ./
RUN go mod download 

# Copy in the source code
EXPOSE 5000 

# Setup an app user so the container doesn't run as the root user
COPY . .
RUN go build -o server .
CMD ["./server"]
