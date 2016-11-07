#!/bin/bash
# -*- coding: utf-8 -*-
# marathon
# https://github.com/topfreegames/marathon
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

VERSION=$(node -e "var pjson = require('./package.json'); console.log(pjson.version);")

docker build -t marathon .
docker login -e="$DOCKER_EMAIL" -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

docker tag marathon:latest tfgco/marathon:$VERSION.$TRAVIS_BUILD_NUMBER
docker tag marathon:latest tfgco/marathon:$VERSION
docker tag marathon:latest tfgco/marathon:latest
docker push tfgco/marathon:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/marathon:$VERSION
docker push tfgco/marathon:latest

DOCKERHUB_LATEST=$(python get_latest_tag.py)

if [ "$DOCKERHUB_LATEST" != "$VERSION.$TRAVIS_BUILD_NUMBER" ]; then
    echo "Last version is not in docker hub!"
    echo "docker hub: $DOCKERHUB_LATEST, expected: $VERSION.$TRAVIS_BUILD_NUMBER"
    exit 1
fi
