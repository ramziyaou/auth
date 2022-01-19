# # start with base image
# FROM mysql:8.0.23

# # import data into container
# # All scripts in docker-entrypoint-initdb.d/ are automatically executed during container startup


# FROM postgres
# # set pgdata to some path outside the VOLUME that is declared in the image
# ENV MYSQLDATA /var/lib/mysql/custom
# # TODO: make sure the postgres user owns PGDATA and has access to an new directories
# USER tester
# COPY ./user/repository/mysql/*.sql /docker-entrypoint-initdb.d/
# COPY custom-entrypoint.sh /usr/local/bin/
# RUN custom-entrypoint.sh postgres
# ENTRYPOINT [ "custom-entrypoint.sh" ]
# CMD [ "postgres" ]


FROM mysql:8.0.23 as builder

ENV MYSQL_DATABASE auth

COPY ./user/repository/mysql/*.sql /docker-entrypoint-initdb.d/

RUN head -n-2 < /usr/local/bin/docker-entrypoint.sh > /usr/local/bin/docker-entrypoint.sh
RUN mkdir -p /var/lib/mysql_tmp
RUN docker-entrypoint.sh mysqld --datadir /var/lib/mysql_tmp

FROM mysql:8.0.23

ENV MYSQL_DATABASE auth
ENV MYSQL_USER tester
ENV MYSQL_PASSWORD secret

COPY --from=builder /var/lib/mysql_tmp /var/lib/mysql_tmp

CMD ["mysqld", "--datadir", "/var/lib/mysql_tmp"]