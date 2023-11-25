FROM --platform=linux/arm64 library/postgres

COPY init.sql /docker-entrypoint-initdb.d/
COPY schema.sql /docker-entrypoint-initdb.d/
RUN printf '\connect queue_share\n' > /docker-entrypoint-initdb.d/qs_schema.sql
RUN cat /docker-entrypoint-initdb.d/schema.sql >> /docker-entrypoint-initdb.d/qs_schema.sql
RUN rm /docker-entrypoint-initdb.d/schema.sql