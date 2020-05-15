#!/usr/bin/bash
sudo docker run -it --rm \
  -v $PWD/mock:/local \
  -v $PWD/docs/swagger:/local/docs/swagger \
  openapitools/openapi-generator-cli:v3.3.4 generate \
  -i /local/docs/swagger/openapi.yml \
  -g go-server \
  -o /local
cd mock
sudo docker build ./ -t trap_collection_mock
sudo docker run --rm trap_collection_mock