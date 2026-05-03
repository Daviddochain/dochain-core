all: build-cosmovisor-linux

build-cosmovisor-linux:
	mkdir -p $(BUILDDIR)
	docker build --platform linux/amd64 --no-cache --tag dochain/dochaind.cosmovisor-binary --file Dockerfile.cosmovisor .
	docker create --platform linux/amd64 --name temp dochain/dochaind.cosmovisor-binary:latest
	docker cp temp:/go/bin/cosmovisor $(BUILDDIR)/
	docker rm temp

.PHONY: all build-cosmovisor-linux





