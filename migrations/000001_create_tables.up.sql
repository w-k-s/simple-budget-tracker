CREATE SCHEMA IF NOT EXISTS budget;

CREATE SEQUENCE IF NOT EXISTS budget.user_id;
ALTER SEQUENCE budget.user_id RESTART WITH 1;
CREATE TABLE IF NOT EXISTS budget.user(
   id BIGINT PRIMARY KEY,
   email VARCHAR (255) UNIQUE NOT NULL,
   created_at TIMESTAMP WITH TIME ZONE NOT NULL,
   created_by VARCHAR (255) NOT NULL,
   last_modified_at TIMESTAMP WITH TIME ZONE,
   last_modified_by VARCHAR (255),
   version BIGINT NOT NULL
);

CREATE SEQUENCE IF NOT EXISTS budget.account_id;
ALTER SEQUENCE budget.account_id RESTART WITH 1;
CREATE TABLE IF NOT EXISTS budget.account(
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(25) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_by VARCHAR (255) NOT NULL,
    last_modified_at TIMESTAMP WITH TIME ZONE,
    last_modified_by VARCHAR (255),
    version BIGINT NOT NULL,
    CONSTRAINT uq_account_name_per_user UNIQUE (user_id, name),
    CONSTRAINT fk_account_user FOREIGN KEY(user_id) REFERENCES budget.user(id) ON DELETE CASCADE
);

CREATE SEQUENCE IF NOT EXISTS budget.category_id;
ALTER SEQUENCE budget.category_id RESTART WITH 1;
CREATE TABLE IF NOT EXISTS budget.category(
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(25) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_by VARCHAR (255) NOT NULL,
    last_modified_at TIMESTAMP WITH TIME ZONE,
    last_modified_by VARCHAR (255),
    last_used_at TIMESTAMP WITH TIME ZONE,
    version BIGINT NOT NULL,
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
    beneficiary_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_by VARCHAR (255) NOT NULL,
    last_modified_at TIMESTAMP WITH TIME ZONE,
    last_modified_by VARCHAR (255),
    version BIGINT NOT NULL,
    CONSTRAINT fk_record_account FOREIGN KEY(account_id) REFERENCES budget.account(id) ON DELETE CASCADE,
    CONSTRAINT fk_record_category FOREIGN KEY(category_id) REFERENCES budget.category(id) ON DELETE NO ACTION,
    CONSTRAINT fk_record_beneficiary_id FOREIGN KEY(beneficiary_id) REFERENCES budget.account(id) ON DELETE CASCADE
);

-- Prevent changing the currency of an account once created

CREATE OR REPLACE FUNCTION fnprevent_update()
  RETURNS trigger AS
$BODY$
    BEGIN
        RAISE EXCEPTION 'Column value can not be updated';
    END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;

DROP TRIGGER IF EXISTS trg_prevent_update ON budget.account;
CREATE TRIGGER trg_prevent_update
  BEFORE UPDATE OF currency
  ON budget.account
  FOR EACH ROW
  EXECUTE PROCEDURE fnprevent_update();

-- Automatically audit rows

create or replace function audit_record()
returns trigger as $body$
  begin
    IF (TG_OP = 'UPDATE') THEN
        new.version = old.version + 1;
        new.last_modified_date = now() at time zone 'utc';
    ELSIF (TG_OP = 'INSERT') THEN
         new.version = 1;
         new.created_date = now() at time zone 'utc';
    END IF;
    return new;
  end
$body$
language plpgsql;

DROP TRIGGER IF EXISTS audit_user ON budget.user;
create trigger audit_user
BEFORE update on budget.user
for each row execute procedure audit_record();

DROP TRIGGER IF EXISTS audit_account ON budget.account;
create trigger audit_account
BEFORE update on budget.account
for each row execute procedure audit_record();

DROP TRIGGER IF EXISTS audit_category ON budget.category;
create trigger audit_category
BEFORE update on budget.category
for each row execute procedure audit_record();

DROP TRIGGER IF EXISTS audit_record ON budget.record;
create trigger audit_record
BEFORE update on budget.record
for each row execute procedure audit_record();