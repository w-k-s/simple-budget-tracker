DROP TRIGGER IF EXISTS trg_prevent_update ON budget.account;
CREATE TRIGGER trg_prevent_update
  BEFORE UPDATE OF currency
  ON budget.account
  FOR EACH ROW
  EXECUTE PROCEDURE fnprevent_update();