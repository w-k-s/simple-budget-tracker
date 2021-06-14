CREATE SCHEMA IF NOT EXISTS budget;

CREATE SEQUENCE IF NOT EXISTS budget.user_id;
CREATE TABLE IF NOT EXISTS budget.user(
   id BIGSERIAL PRIMARY KEY,
   email VARCHAR (255) UNIQUE NOT NULL
);

CREATE SEQUENCE IF NOT EXISTS budget.account_id;
CREATE TABLE IF NOT EXISTS budget.account(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGSERIAL NOT NULL,
    name VARCHAR(255) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    CONSTRAINT fk_account_user FOREIGN KEY(user_id) REFERENCES budget.user(id) ON DELETE CASCADE
);

CREATE SEQUENCE IF NOT EXISTS budget.category_id;
CREATE TYPE budget.category_type AS ENUM ('INCOME', 'EXPENSE');
CREATE TABLE IF NOT EXISTS budget.category(
    id BIGSERIAL PRIMARY KEY,
    account_id BIGSERIAL NOT NULL,
    name VARCHAR(255) NOT NULL,
    type budget.category_type NOT NULL,
    CONSTRAINT fk_category_account FOREIGN KEY(account_id) REFERENCES budget.account(id) ON DELETE CASCADE
);

CREATE SEQUENCE IF NOT EXISTS budget.entry_id;
CREATE TYPE budget.entry_type AS ENUM ('INCOME', 'EXPENSE', 'TRANSFER');
CREATE TABLE IF NOT EXISTS budget.entry(
    id BIGSERIAL PRIMARY KEY,
    account_id BIGSERIAL NOT NULL,
    category_id BIGSERIAL NOT NULL,
    memo VARCHAR(255) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    amount_minor_units DECIMAL(19,0) NOT NULL DEFAULT 0,
    date DATE NOT NULL,
    type budget.entry_type NOT NULL,
    transfer_account_id BIGSERIAL,
    CONSTRAINT fk_entry_account FOREIGN KEY(account_id) REFERENCES budget.account(id) ON DELETE CASCADE,
    CONSTRAINT fk_entry_category FOREIGN KEY(category_id) REFERENCES budget.category(id) ON DELETE NO ACTION,
    CONSTRAINT fk_entry_transfer_account_id FOREIGN KEY(transfer_account_id) REFERENCES budget.account(id) ON DELETE CASCADE
);