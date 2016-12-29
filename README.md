Marathon
========
[![Build Status](https://travis-ci.org/topfreegames/marathon.svg?branch=master)](https://travis-ci.org/topfreegames/marathon)
[![Docs](https://readthedocs.org/projects/marathon/badge/?version=latest
)](http://marathon.readthedocs.io/en/latest/)
[![](https://imagelayers.io/badge/tfgco/marathon:latest.svg)](https://imagelayers.io/?images=tfgco/marathon:latest 'Marathon Image Layers')

The Marathon push notification platform makes it very easy to send massive push notifications to tens of millions of users for several different apps. The architecture is composed of two main modules:
- An API built on top of [Echo Web Framework](https://github.com/labstack/echo).
- Workers built on top of [go-workers](https://github.com/jrallison/go-workers).

## Features

* **Multi-tenant** - Marathon already works for as many apps as you need, just keep adding new ones;
* **Multi-services** - Marathon supports both gcm and apns services, but plugging a new one shouldn't be difficult;
* **Massive Push Notification** - Send tens of millions of push notifications and keep track of job status;
* **New Relic Support** - Natively support new relic with segments in each API route for easy detection of bottlenecks;
* **Sendgrid Support** - Natively support sendgrid and send emails when jobs are created, scheduled, paused or enter circuit break;
* **Easy to deploy** - Marathon comes with containers already exported to docker hub for every single of our successful builds. Just pick your choice!

Read more about Marathon in our [comprehensive documentation](http://marathon.readthedocs.io/).

## Hacking Marathon

### Setup

Make sure you have go installed on your machine.
If you use homebrew you can install it with `brew install go`.

Run `make setup`.

### Running the application

Make sure you have docker compose installed as we'll use it to run all Marathon dependencies:
- Kafka
- Zookeper
- Postgres
- Redis

Create the development database with

```
make create-db
make migrate
```

Run the api with `make run-api`.

Run the workers with `make run-workers`.

### Tests

Running tests can be done with `make test`, while creating the test database can be accomplished with `make drop-test` and `make db-test`.
