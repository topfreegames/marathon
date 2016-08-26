# vim:set ft=dockerfile:
FROM postgres:9.5

ENV PGDATA /pg_data

# RUNNING IN MEMORY ################################################################################
# RUN mkdir $PGDATA
# RUN chmod 777 $PGDATA
# RUN mount -t tmpfs -o size=256M tmpfs $PGDATA
# RUN mount
#
# VOLUME /pg_data
# RUNNING IN MEMORY ################################################################################

RUN rm -rf docker-entrypoint.sh
COPY docker-entrypoint.sh /
COPY createfs.sh /

RUN mkdir $PGDATA
RUN chmod 777 $PGDATA
VOLUME /pg_data


ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 5432
CMD ["postgres"]
