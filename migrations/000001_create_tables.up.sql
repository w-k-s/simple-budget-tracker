CREATE SCHEMA IF NOT EXISTS budget;

CREATE SEQUENCE IF NOT EXISTS budget.user_id;
ALTER SEQUENCE budget.user_id RESTART WITH 1;
CREATE TABLE IF NOT EXISTS budget.user(
   id BIGINT PRIMARY KEY,
   email VARCHAR (255) UNIQUE NOT NULL
);

CREATE SEQUENCE IF NOT EXISTS budget.account_id;
ALTER SEQUENCE budget.account_id RESTART WITH 1;
CREATE TABLE IF NOT EXISTS budget.account(
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(25) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    CONSTRAINT uq_account_name_per_user UNIQUE (user_id, name),
    CONSTRAINT fk_account_user FOREIGN KEY(user_id) REFERENCES budget.user(id) ON DELETE CASCADE
);

CREATE SEQUENCE IF NOT EXISTS budget.category_id;
ALTER SEQUENCE budget.category_id RESTART WITH 1;
CREATE TABLE IF NOT EXISTS budget.category(
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(25) NOT NULL,
    CONSTRAINT uq_category_name_per_user UNIQUE (user_id, name),
    CONSTRAINT fk_category_user FOREIGN KEY(user_id) REFERENCES budget.user(id) ON DELETE CASCADE
);

CREATE SEQUENCE IF NOT EXISTS budget.record_id;
ALTER SEQUENCE budget.record_id RESTART WITH 1;
CREATE TYPE budget.record_type AS ENUM ('INCOME', 'EXPENSE', 'TRANSFER');
CREATE TABLE IF NOT EXISTS budget.record(
    id BIGINT PRIMARY KEY,
    account_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    note VARCHAR(50) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    amount_minor_units DECIMAL(19,0) NOT NULL DEFAULT 0,
    date DATE NOT NULL,
    type budget.record_type NOT NULL,
    transfer_account_id BIGINT,
    CONSTRAINT fk_record_account FOREIGN KEY(account_id) REFERENCES budget.account(id) ON DELETE CASCADE,
    CONSTRAINT fk_record_category FOREIGN KEY(category_id) REFERENCES budget.category(id) ON DELETE NO ACTION,
    CONSTRAINT fk_record_transfer_account_id FOREIGN KEY(transfer_account_id) REFERENCES budget.account(id) ON DELETE CASCADE
);