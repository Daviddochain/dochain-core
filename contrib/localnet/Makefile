all: dochaind-env dochaind-upgrade-env

dochaind-env:
	docker build --platform linux/amd64 --no-cache --build-arg UID=$(shell id -u) --build-arg GID=$(shell id -g) --tag classic-do/dochaind-env dochaind-env

dochaind-upgrade-env:
	docker build --platform linux/amd64 --no-cache --build-arg UID=$(shell id -u) --build-arg GID=$(shell id -g) --build-arg entrypoint=cosmovisor-entrypoint.sh --tag classic-do/dochaind-upgrade-env dochaind-env

.PHONY: all dochaind-env dochaind-upgrade-env





