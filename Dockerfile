FROM postgres:16-alpine3.19
ENV POSTGRES_PASSWORD=postgres
ENV POSTGRES_DB=orchestrator
COPY ./sql/yandex.sql /docker-entrypoint-initdb.d/