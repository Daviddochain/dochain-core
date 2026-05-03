#!/usr/bin/env sh

BINARY=/dochaind/${BINARY:-cosmovisor}
ID=${ID:-0}
LOG=${LOG:-dochaind.log}

if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'dochaind'"
	exit 1
fi

BINARY_CHECK="$(file "$BINARY" | grep 'ELF 64-bit LSB executable, x86-64')"

if [ -z "${BINARY_CHECK}" ]; then
	echo "Binary needs to be OS linux, ARCH amd64"
	exit 1
fi

export DOCHAIND_HOME="/dochaind/node${ID}/dochaind"

if [ -d "$(dirname "${DOCHAIND_HOME}"/"${LOG}")" ]; then
    "${BINARY}" run "$@" --home "${DOCHAIND_HOME}" | tee "${DOCHAIND_HOME}/${LOG}"
else
    "${BINARY}" run "$@" --home "${DOCHAIND_HOME}"
fi





