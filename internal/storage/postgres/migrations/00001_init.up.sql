BEGIN;

CREATE TABLE gauge (
                       id VARCHAR(200) PRIMARY KEY UNIQUE NOT NULL,
                       gauge DOUBLE PRECISION NOT NULL
);

CREATE TABLE counter (
                         id VARCHAR(200) PRIMARY KEY UNIQUE NOT NULL,
                         counter INT NOT NULL,
                         CONSTRAINT counter_positive_check CHECK (counter::numeric > 0)
);

COMMIT;
