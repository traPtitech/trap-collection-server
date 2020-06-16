#!/usr/bin/bash
sudo COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 docker-compose -f docker/mock/docker-compose.yml up
sudo docker-compose -f docker/mock/docker-compose.yml down