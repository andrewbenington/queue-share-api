CREATE USER queue_share WITH LOGIN SUPERUSER PASSWORD 'queue_share';

CREATE DATABASE queue_share;

GRANT ALL PRIVILEGES ON DATABASE queue_share TO queue_share;

