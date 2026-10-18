BEGIN;
DROP TRIGGER IF EXISTS trg_orderflow_sync_mock_provider_terminal_status ON payments;
DROP TRIGGER IF EXISTS trg_orderflow_create_mock_provider_transaction ON payments;
DROP TRIGGER IF EXISTS trg_orderflow_assign_mock_provider_reference ON payments;
DROP FUNCTION IF EXISTS orderflow_sync_mock_provider_terminal_status();
DROP FUNCTION IF EXISTS orderflow_create_mock_provider_transaction();
DROP FUNCTION IF EXISTS orderflow_assign_mock_provider_reference();
DROP TABLE IF EXISTS mock_provider_transactions;
COMMIT;
