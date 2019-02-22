#!/bin/bash

if [ "$TRAVIS_BRANCH" != "master" ]; then
  exit 0
fi

if [ "$LIBRDKAFKA_VERSION" != "master" ]; then
  echo "building image to send to dockerhub"
  VERSION=$(cat ./api/version.go | grep "var VERSION" | awk ' { print $4 } ' | sed s/\"//g)
  COMMIT=$(git rev-parse --short HEAD)

  docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

  docker build -t marathon .
  if [ $? -ne 0 ]; then
      exit 1
  fi

  docker tag marathon:latest tfgco/marathon:$TRAVIS_BUILD_NUMBER-v$VERSION-$COMMIT
  docker tag marathon:latest tfgco/marathon:v$VERSION
  docker tag marathon:latest tfgco/marathon:latest
  docker push tfgco/marathon:$TRAVIS_BUILD_NUMBER-v$VERSION-$COMMIT
  if [ $? -ne 0 ]; then
      exit 1
  fi
  docker push tfgco/marathon:v$VERSION
  if [ $? -ne 0 ]; then
      exit 1
  fi
  docker push tfgco/marathon:latest
  if [ $? -ne 0 ]; then
      exit 1
  fi
fi
