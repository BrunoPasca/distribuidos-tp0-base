#!/bin/bash

a=0

cat <<EOF > "$1"
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net

EOF

until [ "$a" -gt "$2" ]; do
    cat <<EOF >> "$1"
client$a:
  container_name: client$a
  image: client:latest
  entrypoint: /client
  environment:
    - CLI_ID=$a
    - CLI_LOG_LEVEL=DEBUG
  networks:
    - testing_net
  depends_on:
    - server

EOF
    a=$((a + 1))
done

cat <<EOF >> "$1"
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
EOF
