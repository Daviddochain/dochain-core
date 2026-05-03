#!/bin/bash

VERSION="${1:-v0.5.11-oracle}"

pushd .. 

git checkout $VERSION
docker build -t dochain/core:$VERSION .
git checkout -

popd

docker build --build-arg version=$VERSION --build-arg chainid=cookie-1 -t dochain/core-node:$VERSION .
docker build --build-arg version=$VERSION --build-arg chainid=bombay-12 -t dochain/core-node:$VERSION-testnet .





