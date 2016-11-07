# Marathon Push Notification Platform

[![Build Status](https://travis-ci.org/topfreegames/marathon.svg?branch=master)](https://travis-ci.org/topfreegames/marathon)
[![Coverage Status](https://coveralls.io/repos/github/topfreegames/marathon/badge.svg?branch=master)](https://coveralls.io/github/topfreegames/marathon?branch=master)
[![Docs](https://readthedocs.org/projects/marathon/badge/?version=latest
)](http://marathon.readthedocs.io/en/latest/)
[![](https://imagelayers.io/badge/tfgco/marathon:latest.svg)](https://imagelayers.io/?images=tfgco/marathon:latest 'Marathon Image Layers')

The Marathon push notification platform makes it very easy to send massive push notifications to tens of millions of users for several different apps.

## Features

* **Multi-tenant** - Marathon already works for as many apps as you need, just keep adding new ones;
* **Massive Push Notification** - Send tens of millions of push notifications and keep track of job status;
* **New Relic Support** - Natively support new relic with segments in each API route for easy detection of bottlenecks;
* **Easy to deploy** - Marathon comes with containers already exported to docker hub for every single of our successful builds. Just pick your choice!

Read more about Marathon in our [comprehensive documentation](http://marathon.readthedocs.io/).

## Hacking Marathon

### Setup

Make sure you have node.js installed on your machine.
If you use homebrew you can install it with `brew install node`.

Run `make setup`.

### Running the application

Create the development database with `make migrate` (first time only).

Run the api with `make run`.

### Running with docker

Provided you have docker installed, to build Marathon's image run:

    $ make build-docker

To run a new marathon instance, run:

    $ make run-docker

### Tests

Running tests can be done with `make test`, while creating the test database can be accomplished with `make drop-test` and `make db-test`.

### Static Analysis

Marathon goes through some static analysis tools for go. To run them just use `make static`.
