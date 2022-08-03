########################################
# STEP 1 Build the binary
#
FROM golang:1.18-alpine as builder

# Install stuff
RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

ENV USER=docker \
    UID=12345

# Create a non-root user
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "$(pwd)" \
    --no-create-home \
    --uid "$UID" \
    "$USER"

WORKDIR $GOPATH/src/github.com/bokuwazheng/fwdr
COPY . .

# Fetch dependencies using go mod
RUN go mod download
RUN go mod verify

# Build the binary using go build
RUN CGO_ENABLED=0 \
    GOOS=linux GOARCH=arm64 \
    go build -ldflags "-s -w" \
	-o /bin/fwdr ./cmd/fwdr

########################################
# STEP 2 Run the binary
#
FROM scratch

# Copy stuff
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the binary and run it as non-privileged user
COPY --from=builder /bin/fwdr /bin/fwdr
USER docker:docker
CMD [ "/bin/fwdr" ]
