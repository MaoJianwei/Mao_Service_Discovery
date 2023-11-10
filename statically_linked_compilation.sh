set -x
# Build a executable program which is statically linked.

# CGO_ENABLED=0 go build -a -v -ldflags '-extldflags "-static"'
CGO_ENABLED=0 go build -a -v -ldflags "-extldflags '-static' -X main.GIT_VERSION=$(git rev-parse --abbrev-ref HEAD)-$(git rev-parse HEAD)"
set +x
