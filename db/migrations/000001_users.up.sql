BEGIN;

CREATE TABLE users
(
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR NOT NULL,
    "description" TEXT NULL
);

COMMIT;