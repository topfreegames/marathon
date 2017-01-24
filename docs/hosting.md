Hosting Marathon
=================

There are two ways to host Marathon: docker or from source.

## Docker

Running Marathon with docker is rather simple. Our docker container image comes bundled with the API binary. All you need to do is load balance all the containers and you're good to go. The API runs at port `8080` in the docker image.

Marathon uses PostgreSQL to store jobs information. The container takes environment variables to specify this connection:

* `MARATHON_DB_HOST` - PostgreSQL host to connect to;
* `MARATHON_DB_PORT` - PostgreSQL port to connect to;
* `MARATHON_DB_DATABASE` - PostgreSQL database to connect to;
* `MARATHON_DB_USER` - Password of the PostgreSQL Server to connect to;

Marathon also uses another PostgreSQL database to read tokens information. The container takes environment variables to specify this connection:

* `MARATHON_PUSH_DB_HOST` - PostgreSQL host to connect to;
* `MARATHON_PUSH_DB_PORT` - PostgreSQL port to connect to;
* `MARATHON_PUSH_DB_DATABASE` - PostgreSQL database to connect to;
* `MARATHON_PUSH_DB_USER` - Password of the PostgreSQL Server to connect to;

For uploading and reading CSV files Marathon uses AWS S3, so you'll need to specify the following environment variables as well:

* `MARATHON_S3_BUCKET` - AWS S3 bucket containing the csv files;
* `MARATHON_S3_FOLDER` - AWS S3 folder containing the csv files;
* `MARATHON_S3_ACCESSKEY` - AWS S3 access key;
* `MARATHON_S3_SECRETACCESSKEY` - AWS S3 secret;

The workers use redis for queueing:

* `MARATHON_WORKERS_REDIS_HOST` - Redis host to connect to;
* `MARATHON_WORKERS_REDIS_PORT` - Redis port to connect to;
* `MARATHON_WORKERS_REDIS_PASS` - Password of the redis server to connect to;

The workers use kafka and zookeper for sending push notifications:

* `MARATHON_WORKERS_ZOOKEEPER_HOSTS` - Zookeeper hosts to connect to;
* `MARATHON_WORKERS_TOPICTEMPLATE` - Kafka topic template;
* `MARATHON_WORKERS_ZOOKEEPER_PREFIX` - The prefix that contains kafka brokers info; e.g. /prefix for accessing brokers node at /prefix/brokers

Finally, the feedback listener uses kafka for receiving the push notifications' feedbacks from APNS or GCM:

* `MARATHON_FEEDBACKLISTENER_KAFKA_BROKERS` - Kafka brokers to connect to (comma separated, without spaces);
* `MARATHON_FEEDBACKLISTENER_KAFKA_TOPICS` - Array of kafka topics to read from;
* `MARATHON_FEEDBACKLISTENER_KAFKA_GROUP` - Kafka consumer group;
* `MARATHON_FEEDBACKLISTENER_FLUSHINTERVAL` - Interval during which the feedback listener caches the feedbacks metrics before updating the job feedbacks in PostgreSQL;

Other than that, there are a couple more configurations you can pass using environment variables:

* `MARATHON_NEWRELIC_KEY` - If you have a [New Relic](https://newrelic.com/) account, you can use this variable to specify your API Key to populate data with New Relic API;
* `MARATHON_SENTRY_URL` - If you have a [sentry server](https://docs.getsentry.com/hosted/) you can use this variable to specify your project's URL to send errors to.
* `MARATHON_SENDGRID_KEY` - If you have a [sendgrid](https://sendgrid.com/) account, you can use this variable to specify your API Key for sending emails when jobs are created, scheduled, paused or enter circuit break;

### Example command for running with Docker

```
    $ docker pull tfgco/marathon
    $ docker run -t --rm -e "MARATHON_POSTGRES_HOST=<postgres host>" -e "MARATHON_POSTGRES_PORT=<postgres port>" -p 8080:8080 tfgco/marathon
```

## Source

Left as an exercise to the reader.
