CREATE DATABASE acid_demo;
\c acid_demo;

CREATE TABLE accounts(
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    balance NUMERIC(10,2) NOT NULL CHECK(balance >= 0)
);

INSERT INTO accounts(name,balance) VALUES('ALICE',500.00),('Bob',200.00);