#!/bin/bash

echo "Getting latest version of the code from git"
git pull origin production

# terminate the docker container
echo "Terminating the docker container"
docker-compose down

echo "Building the docker container"
docker-compose up -d --build
