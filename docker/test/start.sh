#!/bin/bash
cd ./docker/test

stop () {
  docker-compose -p marathon_test stop
  docker-compose -p marathon_test rm -f -v
}

stop

docker network create marathon_test

echo "MY IP IS:" $MY_IP

env MY_IP=${MY_IP} docker-compose -p marathon_test up --remove-orphans -d
env MY_IP=${MY_IP} docker-compose -p marathon_test scale postgres=1 redis=1 zookeeper=1 kafka=1

CHECK_KAFKA=$(docker exec marathontest_kafka_1 /opt/kafka_2.11-0.9.0.0/bin/kafka-topics.sh --list --zookeeper zookeeper:2181 | wc -l )

# FIXME: Try do automate the number of topics
TOTAL_TOPICS=26
until [ $CHECK_KAFKA -ge $TOTAL_TOPICS ]; do
  CHECK_KAFKA=$(docker exec marathontest_kafka_1 /opt/kafka_2.11-0.9.0.0/bin/kafka-topics.sh --list --zookeeper zookeeper:2181 | wc -l )
  echo 'Waiting for Kafka:' $CHECK_KAFKA 'topics created of' $TOTAL_TOPICS '...' && sleep 0.5
done

until docker exec marathontest_postgres_1 pg_isready
  do echo 'Waiting for Postgres...' && sleep 0.5
done

createuser -h localhost -p 9910 -s -U postgres marathon
dropdb -h localhost -p 9910 -U marathon marathon
createdb -h localhost -p 9910 -U marathon marathon

cd ./../..

./marathon migrate -c ./config/test.yaml
