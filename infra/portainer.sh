#!/bin/bash

docker run -d \
    --name portainer \
    --network vpc \
    -p 8000:8000 \
    -p 9443:9443 \
    --restart=always \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v portainer_data:/data \
    portainer/portainer-ce:lts