# marathon
# https://github.com/topfreegames/marathon
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

FROM node:6.8.0

RUN apt-get update -y
RUN apt-get install -yq \
      net-tools \
			build-essential \
			git \
			ca-certificates \
			unzip -y

RUN mkdir -p /var/apps/marathon/lib
RUN mkdir -p /var/apps/marathon/config
WORKDIR /var/apps/marathon
COPY package.json /var/apps/marathon
RUN npm install --silent --no-progress
COPY Makefile /var/apps/marathon

RUN make setup-global

COPY ./lib /var/apps/marathon/lib
COPY ./config /var/apps/marathon/config

ENV NODE_ENV production
ENV PORT 8080
ENV LOG_LEVEL warn
ENV REDIS_URL redis://localhost:22223
ENV PG_URL postgresql://marathon@localhost:22222/marathon
ENV KAFKA_URL localhost:22224
ENV KAFKA_CLIENT_ID marathonApiProducer

EXPOSE 8080

CMD ["node", "lib/marathon.js", "start"]
