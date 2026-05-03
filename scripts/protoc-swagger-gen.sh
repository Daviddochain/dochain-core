#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR"

mkdir -p ./tmp-swagger-gen
mkdir -p ./client/docs/swagger-ui

if ! command -v buf >/dev/null 2>&1; then
  echo "ERROR: buf is not installed or not in PATH."
  echo "Install buf before generating swagger."
  exit 1
fi

if ! command -v npx >/dev/null 2>&1; then
  echo "ERROR: npx is not installed or not in PATH."
  echo "Install Node.js/npm before generating swagger."
  exit 1
fi

cosmos_sdk_dir="$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk)"
ibc_go_dir="$(go list -f '{{ .Dir }}' -m github.com/cosmos/ibc-go/v10)"
wasm_dir="$(go list -f '{{ .Dir }}' -m github.com/CosmWasm/wasmd)"

pushd proto >/dev/null

proto_dirs=$(
  find ./ "$cosmos_sdk_dir/proto" "$ibc_go_dir/proto" "$wasm_dir/proto" \
    -name '*.proto' -print0 |
    xargs -0 -n1 dirname |
    sort |
    uniq
)

for dir in $proto_dirs; do
  query_files=$(find "$dir" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \) -print)

  if [[ -n "$query_files" ]]; then
    while IFS= read -r query_file; do
      [[ -n "$query_file" ]] && buf generate --template buf.gen.swagger.yaml "$query_file"
    done <<< "$query_files"
  fi
done

popd >/dev/null

npx swagger-combine ./client/docs/config.json \
  -o ./client/docs/swagger-ui/swagger.json \
  --continueOnConflictingPaths true \
  --includeDefinitions true

cp ./client/docs/swagger-ui/swagger.json ./client/docs/swagger-ui/swagger.yaml

rm -rf ./tmp-swagger-gen

echo "Swagger generated: ./client/docs/swagger-ui/swagger.yaml"