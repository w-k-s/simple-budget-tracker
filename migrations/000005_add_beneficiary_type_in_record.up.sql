ALTER TABLE budget.record 
ADD COLUMN beneficiary_type VARCHAR(30);

UPDATE budget.record
SET beneficiary_type = 'Current';
