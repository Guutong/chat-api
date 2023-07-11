# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# The image contains all the necessary build tools and libraries needed for Go.
FROM golang:1.17-buster as builder

# Create and change to the app directory.
WORKDIR /app

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN go build -v -o server

# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:buster-slim

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /server

# Run the web service on container startup.
CMD ["/server"]