FROM postgres:12-alpine


ARG POSTGRES_HOST
ARG POSTGRES_PORT
ARG POSTGRES_USER
ARG POSTGRES_DB
ARG PGPASSWORD

COPY seed.sql ./seed.sql
COPY init.sh ./init.sh

RUN chmod +x ./init.sh

CMD ./init.sh
