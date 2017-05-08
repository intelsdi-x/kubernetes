FROM golang:latest

# Copy the local package files to the container’s workspace.
ADD . /go/src/github.com/user/sample-hugepage-app

RUN go get github.com/golang/glog
# Build the golang-docker command inside the container.
RUN go install github.com/user/sample-hugepage-app

# Run the golang-docker command when the container starts.
CMD exec /bin/bash -c "trap : TERM INT; sleep infinity & wait" 
