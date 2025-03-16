#!/bin/bash
# Usage: ./generar-compose.sh <filename> <num_clients>
# Description: Generates a docker-compose file with a server and num_clients clients

client_id=1
num_clients=$2
filename=$1

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <filename> <num_clients>"
    exit 1
fi

if ! [[ "$num_clients" =~ ^[0-9]+$ ]]; then
    echo "Error: num_clients must be a number"
    exit 1
fi

cat <<EOF > "$filename"
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    volumes:
      - ./server/config.ini:/config.ini
    networks:
      - testing_net

EOF

until [ "$client_id" -gt "$num_clients" ]; do
    cat <<EOF >> "$filename"
  client$client_id:
    container_name: client$client_id
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=$client_id
      - CLI_LOG_LEVEL=DEBUG
    volumes:
      - ./client/config.yaml:/config.yaml
    networks:
      - testing_net
    depends_on:
      - server

EOF
    client_id=$((client_id + 1))
done

cat <<EOF >> "$filename"
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
EOF
