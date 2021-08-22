ALTER TABLE budget.record
ADD COLUMN source_account_id BIGINT REFERENCES budget.account(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE budget.record
ADD COLUMN transfer_reference VARCHAR(128);