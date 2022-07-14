# Appliance Build Test Suite

This directory contains a test suite for testing the setup scripts used to build the VEBA appliance.

This allows changes and additions to these scripts to be tested rapidly rather than having to build the full appliance to validate changes, thus reducing the feedback cycle.

## Prerequisites
You will need the following to run these tests
- A linux machine to run the tests on with this repo cloned
- docker installed

## Image build
The tests run using a docker container to simulate the VEBA appliance environment. Rather than pulling a pre-built image from an image registry you need to build the image locally to ensure it contains the latest code base. A helper script is provided to successfully build the image:
```bash
./build-image.sh
```
Note that only the contents of the [`/files`](/files/) directory from this repo is mounted in to the container when running the tests. As such, changes to any other appliance build code (such as the contents of the [`/scripts`](/scripts/) directory) will require you to re-build the image as above.

## Run tests
The full test suite can be run using the helper script:
```bash
./run-tests.sh
```
The test files are all contained in the [`bats-tests`](bats-tests/) directory and are written using the [bats](https://bats-core.readthedocs.io/en/stable/index.html) testing framework for Bash.
If you need to inspect the running container after running the tests (to help debug failures) you can run the same command with a `-d` switch:
```bash
./run-tests.sh -d
```
This will run the tests again and then drop you to a command line inside the container before it terminates, where you can inspect the content of the configured appliance.