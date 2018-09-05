#!/bin/bash

if ([ "$TRAVIS_BRANCH" == "master" ] || [ ! -z "$TRAVIS_TAG" ]) && [ "$TRAVIS_PULL_REQUEST" == "false" ]; then 
  cd $GOPATH/src/github.com/hashicorp/faas-nomad;
  docker login -u "$DOCKER_USERNAME" -p "$DOCKER_PASSWORD" quay.io; 
  goreleaser --rm-dist --skip-validate
fi
