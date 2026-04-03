all: build-cosmovisor-linux

build-cosmovisor-linux:
	mkdir -p $(BUILDDIR)
	docker build --platform linux/amd64 --no-cache --tag classic-do/terraclassic.cosmovisor-binary --file Dockerfile.cosmovisor .
	docker create --platform linux/amd64 --name temp classic-do/terraclassic.cosmovisor-binary:latest
	docker cp temp:/go/bin/cosmovisor $(BUILDDIR)/
	docker rm temp

.PHONY: all build-cosmovisor-linux




