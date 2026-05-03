#!/usr/bin/env sh

set -e
export PATH="$PATH:/go/bin"

echo "Generating gogo proto code"
cd proto
proto_dirs=$(find do -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep go_package "$file" >/dev/null 2>&1; then
      buf generate --template buf.gen.gogo.yaml "$file"
    fi
  done
done