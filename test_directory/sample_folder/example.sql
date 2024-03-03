CREATE DATABASE IF NOT EXISTS EXERCISE_DB;

CREATE OR REPLACE TABLE CUSTOMERS (
    ID INT,
    first_name varchar,
    last_name varchar,
    email varchar,
    age INT,
    city varchar
);

-- COPY INTO EXERCISE_DB.PUBLIC.CUSTOMERS
-- FROM s3://snowflake-assignments-mc/gettingstarted/customers.csv
-- FILE_FORMAT = (
--     type = csv
--     field_delimiter = ','
--     skip_header = 1
-- );

USE DATABASE MANAGE_DB;
USE SCHEMA EXTERNAL_STAGES;

CREATE OR REPLACE STAGE aws_assignment_stage
    url = 's3://snowflake-assignments-mc/loadingdata/'
    file_format = (
        type = csv
        field_delimiter = ';'
        skip_header = 1
    );

DESC STAGE MANAGE_DB.EXTERNAL_STAGES.AWS_ASSIGNMENT_STAGE;
LIST @MANAGE_DB.EXTERNAL_STAGES.AWS_ASSIGNMENT_STAGE;

COPY INTO EXERCISE_DB.PUBLIC.CUSTOMERS
FROM @MANAGE_DB.EXTERNAL_STAGES.AWS_ASSIGNMENT_STAGE
    pattern = '.*customers.*\.csv$';

SELECT * FROM EXERCISE_DB.PUBLIC.CUSTOMERS LIMIT 100;
SELECT COUNT(*) FROM EXERCISE_DB.PUBLIC.CUSTOMERS;