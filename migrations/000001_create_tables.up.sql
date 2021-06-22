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
    name VARCHAR(255) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    CONSTRAINT uq_account_name_per_user UNIQUE (user_id, name),
    CONSTRAINT fk_account_user FOREIGN KEY(user_id) REFERENCES budget.user(id) ON DELETE CASCADE
);

CREATE SEQUENCE IF NOT EXISTS budget.category_id;
ALTER SEQUENCE budget.category_id RESTART WITH 1;
CREATE TYPE budget.category_type AS ENUM ('INCOME', 'EXPENSE');
CREATE TABLE IF NOT EXISTS budget.category(
    id BIGINT PRIMARY KEY,
    account_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    type budget.category_type NOT NULL,
    CONSTRAINT uq_category_name_per_account_and_type UNIQUE (account_id, name, type),
    CONSTRAINT fk_category_account FOREIGN KEY(account_id) REFERENCES budget.account(id) ON DELETE CASCADE
);

CREATE SEQUENCE IF NOT EXISTS budget.entry_id;
ALTER SEQUENCE budget.entry_id RESTART WITH 1;
CREATE TYPE budget.entry_type AS ENUM ('INCOME', 'EXPENSE', 'TRANSFER');
CREATE TABLE IF NOT EXISTS budget.entry(
    id BIGINT PRIMARY KEY,
    account_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    memo VARCHAR(255) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    amount_minor_units DECIMAL(19,0) NOT NULL DEFAULT 0,
    date DATE NOT NULL,
    type budget.entry_type NOT NULL,
    transfer_account_id BIGINT,
    CONSTRAINT fk_entry_account FOREIGN KEY(account_id) REFERENCES budget.account(id) ON DELETE CASCADE,
    CONSTRAINT fk_entry_category FOREIGN KEY(category_id) REFERENCES budget.category(id) ON DELETE NO ACTION,
    CONSTRAINT fk_entry_transfer_account_id FOREIGN KEY(transfer_account_id) REFERENCES budget.account(id) ON DELETE CASCADE
);