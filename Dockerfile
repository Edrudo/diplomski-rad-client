FROM golang:1.21.3

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY . .
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./

# Build
RUN go build -o /http3-client

EXPOSE 4242
# Run
CMD ["/http3-client", "https://78.1.126.230:4242/imagePart", "big_auto.jpg"]
