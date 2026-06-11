FROM postgres:18.3-alpine

# Устанавливаем postgis
RUN apk update && apk add --no-cache postgis

# Переносим файлы расширения туда, где их ожидает увидеть PostgreSQL Docker-образа
RUN cp -r /usr/share/postgresql/extension/* /usr/local/share/postgresql/extension/ && \
    cp -r /usr/lib/postgresql/* /usr/local/lib/postgresql/

# Создаем скрипт автоматической активации
RUN echo "CREATE EXTENSION IF NOT EXISTS postgis;" > /docker-entrypoint-initdb.d/init-postgis.sql && \
    chown postgres:postgres /docker-entrypoint-initdb.d/init-postgis.sql
