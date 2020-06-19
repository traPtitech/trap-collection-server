#!/usr/bin/bash
sudo docker run -it --rm \
  -v $PWD/mock:/local \
  -v $PWD/docs/swagger:/local/docs/swagger \
  openapitools/openapi-generator-cli:v3.3.4 generate \
  -i /local/docs/swagger/openapi.yml \
  -g go-server \
  -o /local
sudo docker-compose -f docker/mock/docker-compose.yml up --build
sudo docker-compose -f docker/mock/docker-compose.yml down