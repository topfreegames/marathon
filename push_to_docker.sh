#!/bin/bash

VERSION=$(cat ./api/version.go | grep "var VERSION" | awk ' { print $4 } ' | sed s/\"//g)
COMMIT=$(git rev-parse HEAD)

docker login -e="$DOCKER_EMAIL" -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

docker build -t marathon .
docker tag marathon:latest tfgco/marathon:$TRAVIS_BUILD_NUMBER-v$VERSION-$COMMIT
docker tag marathon:latest tfgco/marathon:$VERSION
docker tag marathon:latest tfgco/marathon:latest
docker push tfgco/marathon:$TRAVIS_BUILD_NUMBER-v$VERSION-$COMMIT
docker push tfgco/marathon:$VERSION
docker push tfgco/marathon:latest
