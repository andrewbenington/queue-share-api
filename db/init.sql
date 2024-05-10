DROP DATABASE IF EXISTS queue_share;

DO
$$BEGIN
IF EXISTS (SELECT FROM pg_roles WHERE rolname = 'queue_share') THEN
    EXECUTE 'REVOKE CONNECT ON DATABASE "postgres" FROM queue_share';
    EXECUTE 'DROP USER queue_share';
END IF;
END$$;
CREATE USER queue_share WITH LOGIN SUPERUSER PASSWORD 'queue_share';

CREATE DATABASE queue_share OWNER queue_share;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO queue_share;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO queue_share;