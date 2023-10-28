#!/bin/bash

echo "open http://localhost:15672/ -- login with guest:guest"

docker run --rm --hostname my-rabbit --name some-rabbit --network host rabbitmq:3-management
