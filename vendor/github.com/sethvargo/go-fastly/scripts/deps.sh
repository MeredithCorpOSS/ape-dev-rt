#!/usr/bin/env bash
#
# This script updates dependencies using a temporary directory. This is required
# to avoid any auxillary dependencies that sneak into GOPATH.
set -e

# Get the name from the command line
NAME=$1
if [ -z $NAME ]; then
  echo "Please specify a name."
  exit 1
fi

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$(cd -P "$(dirname "$SOURCE")/.." && pwd)"

# Change into that directory
cd "$DIR"

# Announce
echo "==> Updating dependencies..."

echo "--> Making tmpdir..."
tmpdir=$(mktemp -d)
function cleanup {
  rm -rf "${tmpdir}"
}
trap cleanup EXIT

export GOPATH="${tmpdir}"
export PATH="${tmpdir}/bin:$PATH"

mkdir -p "${tmpdir}/src/github.com/sethvargo"
pushd "${tmpdir}/src/github.com/sethvargo" &>/dev/null

echo "--> Copying ${NAME}..."
cp -R "$DIR" "${tmpdir}/src/github.com/sethvargo/${NAME}"
pushd "${tmpdir}/src/github.com/sethvargo/${NAME}" &>/dev/null
rm -rf vendor/

echo "--> Installing dependency manager..."
go get -u github.com/kardianos/govendor
govendor init

echo "--> Installing all dependencies (may take some time)..."
govendor fetch -v +outside

echo "--> Vendoring..."
govendor add +external

echo "--> Moving into place..."
vpath="${tmpdir}/src/github.com/sethvargo/${NAME}/vendor"
popd &>/dev/null
popd &>/dev/null
rm -rf vendor/
cp -R "${vpath}" .
