#!/bin/bash

# Define variables
TEST_MESSAGE="Hello Echo Server"
SERVER_CONTAINER="server"
NETCAT_CONTAINER="netcat"
SERVER_PORT="12345"
TIMEOUT=5

result=$(docker run --network tp0_testing_net --rm alpine:latest sh -c "echo '${TEST_MESSAGE}' | nc -w ${TIMEOUT} ${SERVER_CONTAINER} ${SERVER_PORT}" )

if [ "$result" = "$TEST_MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
    exit 0
else
    echo "action: test_echo_server | result: fail"
    exit 1
fi