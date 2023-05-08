#!/usr/bin/env bash
RUN_NAME="GateWay"

mkdir -p output/bin

# Build C++ source file to create dynamic library
#g++ -std=c++11 -shared -o output/lib/libadd.so cgo/test/add.cc

if [ "$IS_SYSTEM_TEST_ENV" != "1" ]; then
    go build -o output/bin/${RUN_NAME}
else
    go test -c -covermode=set -o output/bin/${RUN_NAME} -coverpkg=./...
fi

# Set the environment variable to load the dynamic library
#export LD_LIBRARY_PATH=.

