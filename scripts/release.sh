#!/bin/bash
DIR=$(cd "$(dirname "$0")"/../ && pwd)

LATEST_TAG_NAME=$(git describe --tags --exact-match HEAD 2>/dev/null)
if [ $? != 0 ]; then
  echo "There is no tag pointing to HEAD"
  exit 1
fi

TAG_HASH=$(git rev-parse --verify $LATEST_TAG_NAME --short)

echo "Releasing $LATEST_TAG_NAME ($TAG_HASH)..."

VERSION_OUTPUT=$(./bin/ape-dev-rt version)
VERSION=$(echo $VERSION_OUTPUT | awk -F\( '{ print $1 }' | awk '{ print $NF }')

if [ "v${VERSION}" != "${LATEST_TAG_NAME}" ]; then
	echo "Ooops, version & tag name don't match (v${VERSION} != ${LATEST_TAG_NAME})"
	exit 1
fi

rm -rf $DIR/release
mkdir -p $DIR/release

echo "Zipping artifacts..."

for PLATFORM in $(find ./pkg -mindepth 1 -maxdepth 1 -type d); do
    OSARCH=$(basename ${PLATFORM})
    FILENAME=$DIR/release/ape-dev-rt_${OSARCH}_${LATEST_TAG_NAME}.zip

    pushd $PLATFORM >/dev/null 2>&1
    echo "--> Zipping $FILENAME"
    zip -j -r $FILENAME ./*
    popd >/dev/null 2>&1
done

echo "Calculating SHA256 of the artifact for darwin..."
SHASUM=$(shasum -a256 $DIR/release/ape-dev-rt_darwin_amd64_${LATEST_TAG_NAME}.zip 2>&1)

echo "Uploading artifacts to S3..."
for ARTIFACT in $(find $DIR/release -mindepth 1 -maxdepth 1 -type f); do
    FILENAME=$(basename $ARTIFACT)
    aws --region us-east-1 s3 cp $ARTIFACT s3://ti-rt-download/${FILENAME}
done

echo
echo $SHASUM
echo
