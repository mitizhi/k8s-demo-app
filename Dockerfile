# Stage 1: Build the Go app
FROM docker.io/library/golang:1.23-alpine AS builder
#FROM docker.io/library/golang:1.23 AS builder

# Set the Current Working Directory inside the container
WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum ./
# Copy the source code into the container
COPY internal internal
COPY app app

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download && go mod verify

# Build the Go app with static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o pseudo-web-app app/main.go

# Prepare the files for the final image
RUN mkdir -pv /staging/bin /staging/data /staging/state
COPY data  /staging/data
RUN mkdir -pv /staging/state && printf "0" > /staging/state/count
RUN cp -aiv pseudo-web-app /staging/bin
RUN ls -lFR /staging

#############################################################################
# Stage 2: Create a minimal runtime image using scratch
FROM scratch

# Copy the pre-built binary file from the builder stage
COPY --from=builder /staging /

# Command to run the executable
ENTRYPOINT ["/bin/pseudo-web-app"]
