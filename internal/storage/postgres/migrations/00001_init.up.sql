BEGIN;

CREATE TABLE gauges (
                       id VARCHAR(200) PRIMARY KEY UNIQUE NOT NULL,
                       gauge DOUBLE PRECISION NOT NULL
);

CREATE TABLE counters (
                         id VARCHAR(200) PRIMARY KEY UNIQUE NOT NULL,
                         counter INT NOT NULL,
                         CONSTRAINT counter_positive_check CHECK (counter::numeric > 0)
);

COMMIT;
