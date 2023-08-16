set -x
# Build a executable problem which is statically linked.
CGO_ENABLED=0 go build -a -v -ldflags '-extldflags "-static"'
set +x