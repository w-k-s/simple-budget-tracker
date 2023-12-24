CREATE TABLE IF NOT EXISTS budget.budget(
   id BIGINT NOT NULL,
   period VARCHAR(20) NOT NULL,
   created_at TIMESTAMP WITH TIME ZONE NOT NULL,
   created_by VARCHAR (255) NOT NULL,
   last_modified_at TIMESTAMP WITH TIME ZONE,
   last_modified_by VARCHAR (255),
   version BIGINT NOT NULL,
   CONSTRAINT pk_budget_id PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS budget.budget_per_category(
    budget_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    amount_minor_units BIGINT NOT NULL,
    CONSTRAINT fk_budget_per_category_budget_id FOREIGN KEY(budget_id) REFERENCES budget.budget(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE,
    CONSTRAINT fk_budget_per_category_category_id FOREIGN KEY(category_id) REFERENCES budget.category(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE,
    CONSTRAINT uq_budget_per_category_category UNIQUE(category_id)
);

CREATE TABLE IF NOT EXISTS budget.account_budgets(
    account_id BIGINT NOT NULL,
    budget_id BIGINT NOT NULL,
    CONSTRAINT fk_account_budgets_account_id FOREIGN KEY(account_id) REFERENCES budget.account(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE,
    CONSTRAINT fk_account_budgets_budget_id FOREIGN KEY(budget_id) REFERENCES budget.budget(id)
        ON UPDATE CASCADE,
        ON DELETE CASCADE
);

DROP TRIGGER IF EXISTS audit_user ON budget.budget;
create trigger audit_user
BEFORE update on budget.budget
for each row execute procedure audit_record();