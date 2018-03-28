APP="carp"
VERSION="0.1.0"

TARGETDIR="target"
PKG="${APP}-${VERSION}"
BINARY="${TARGETDIR}/${APP}"

setup:
	dep ensure

$(BINARY):
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-X main.Version=${VERSION} -extldflags "-static"' -o $(BINARY) .

build: $(BINARY)

clean:
	rm -rf $(TARGETDIR)
