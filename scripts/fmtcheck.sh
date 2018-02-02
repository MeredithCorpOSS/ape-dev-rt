#!/usr/bin/env bash

echo "==> Checking that code complies with gofmt requirements..."
gofmt_files=$(gofmt -l . | grep -v vendor)
if [[ -n ${gofmt_files} ]]; then
    echo 'gofmt needs running on the following files:'
    echo "${gofmt_files}"
    echo "You can use \`gofmt -w .\` to reformat code."
    exit 1
fi

exit 0
