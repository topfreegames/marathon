#
# Copyright (c) 2016 TFG Co <backend@tfgco.com>
# Author: TFG Co <backend@tfgco.com>
#
# Permission is hereby granted, free of charge, to any person obtaining a copy of
# this software and associated documentation files (the "Software"), to deal in
# the Software without restriction, including without limitation the rights to
# use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
# the Software, and to permit persons to whom the Software is furnished to do so,
# subject to the following conditions:
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
# FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
# COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
# IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
# CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
#

FROM golang:1.12-alpine

MAINTAINER TFG Co <backend@tfgco.com>

RUN apk update
RUN apk add make git g++ bash python wget pkgconfig

ENV LIBRDKAFKA_VERSION 0.11.6
RUN wget -O /root/librdkafka-${LIBRDKAFKA_VERSION}.tar.gz https://github.com/edenhill/librdkafka/archive/v${LIBRDKAFKA_VERSION}.tar.gz && \
    tar -xzf /root/librdkafka-${LIBRDKAFKA_VERSION}.tar.gz -C /root && \
    cd /root/librdkafka-${LIBRDKAFKA_VERSION} && \
    ./configure && make && make install && make clean && ./configure --clean && \
    go get -u github.com/golang/dep/cmd/dep

RUN mkdir -p /go/src/github.com/topfreegames/marathon
WORKDIR /go/src/github.com/topfreegames/marathon

ADD Gopkg.lock /go/src/github.com/topfreegames/marathon
ADD Gopkg.toml /go/src/github.com/topfreegames/marathon
RUN dep ensure -vendor-only

ADD . /go/src/github.com/topfreegames/marathon

ENV CPLUS_INCLUDE_PATH /usr/local/include
ENV LIBRARY_PATH /usr/local/lib
ENV LD_LIBRARY_PATH /usr/local/lib
RUN export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:$PKG_CONFIG_PATH && make build

RUN mkdir /app
RUN mv /go/src/github.com/topfreegames/marathon/bin/marathon /app/marathon
RUN mv /go/src/github.com/topfreegames/marathon/config /app/config
RUN rm -r /go/src/github.com/topfreegames/marathon

WORKDIR /app

EXPOSE 8080 8081
VOLUME /app/config

CMD /app/marathon start-api -c /app/config/default.yaml
