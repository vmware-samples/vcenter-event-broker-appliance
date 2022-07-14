#!/bin/bash +x

rootdir="$(git rev-parse --show-toplevel)"
if [ "$1" == "-d" ]; then
  cmd="/bin/bash"
else
  cmd="bats setup/test/bats-tests/"
fi
docker run -it -v "${rootdir}"/files:/root/setup --privileged veba-test ${cmd}
