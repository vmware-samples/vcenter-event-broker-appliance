#!/bin/bash

rootdir="$(git rev-parse --show-toplevel)"
cmd="bats setup/test/bats-tests/"
if [ "$1" == "-d" ]; then
  cmd="${cmd}; /bin/bash"
fi
docker run -it -v "${rootdir}"/files:/root/setup:ro --privileged veba-test /bin/bash -c "${cmd}"
