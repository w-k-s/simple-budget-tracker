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