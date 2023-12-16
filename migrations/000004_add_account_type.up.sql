ALTER TABLE budget.account 
ADD COLUMN account_type VARCHAR(30);

UPDATE budget.account SET account_type = 'Current';

ALTER TABLE budget.account 
ALTER COLUMN account_type SET NOT NULL;