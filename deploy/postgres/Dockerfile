FROM postgres:18.3-alpine

RUN apk update && \
    apk add --no-cache postgis

RUN echo "CREATE EXTENSION IF NOT EXISTS postgis;" > /docker-entrypoint-initdb.d/init-postgis.sql
