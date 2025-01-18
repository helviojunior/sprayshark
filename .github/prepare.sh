#/bin/bash

echo "${{ github.event.release.tag_name }}" | grep -oE '[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' > /tmp/version.txt
git rev-parse --short HEAD > /tmp/githash.txt
go version | cut -d' ' -f 3,4 | sed 's/ /_/g' > /tmp/buildenv.txt
date -u +'%Y-%m-%dT%H:%M:%SZ' > /tmp/buildtime.txt

ver=$(cat /tmp/version.txt)
hash=$(cat /tmp/githash.txt)
buildenv=$(cat /tmp/buildenv.txt)
buildtime=$(cat /tmp/buildtime.txt)

echo "VER: $ver"
echo "Build Hash: $hash"
echo "Build Env: $buildenv"
echo "Build Time: $buildtime"