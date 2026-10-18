BEGIN;

-- Keep simulated provider state separate from OrderFlow's internal payment state.
-- This lets the reconciliation worker recover a terminal provider result when
-- the callback delivery was missed, without pretending the mock provider is live.
CREATE TABLE IF NOT EXISTS mock_provider_transactions (
    payment_id UUID PRIMARY KEY REFERENCES payments(id) ON DELETE CASCADE,
    provider_payment_id TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED')),
    amount_paise BIGINT NOT NULL CHECK (amount_paise >= 0),
    currency CHAR(3) NOT NULL DEFAULT 'INR'
        CHECK (currency ~ '^[A-Z]{3}$'),
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Backfill pre-existing mock payments idempotently. Existing real/provider types
-- are not touched by this mock-only table.
UPDATE payments
SET provider_payment_id = 'mock_' || id::text
WHERE provider = 'MOCK' AND (provider_payment_id IS NULL OR btrim(provider_payment_id) = '');

INSERT INTO mock_provider_transactions (
    payment_id, provider_payment_id, status, amount_paise, currency,
    failure_reason, created_at, updated_at
)
SELECT p.id, p.provider_payment_id,
       CASE WHEN p.status IN ('SUCCESS', 'FAILED') THEN p.status ELSE 'PENDING' END,
       p.amount_paise, p.currency, p.failure_reason, p.created_at, p.updated_at
FROM payments p
WHERE p.provider = 'MOCK'
  AND p.provider_payment_id IS NOT NULL
ON CONFLICT (payment_id) DO NOTHING;

CREATE OR REPLACE FUNCTION orderflow_assign_mock_provider_reference()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.provider = 'MOCK' AND
       (NEW.provider_payment_id IS NULL OR btrim(NEW.provider_payment_id) = '') THEN
        NEW.provider_payment_id := 'mock_' || NEW.id::text;
    END IF;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION orderflow_create_mock_provider_transaction()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.provider = 'MOCK' THEN
        INSERT INTO mock_provider_transactions (
            payment_id, provider_payment_id, status, amount_paise, currency,
            failure_reason, created_at, updated_at
        ) VALUES (
            NEW.id, NEW.provider_payment_id, 'PENDING', NEW.amount_paise,
            NEW.currency, NULL, NEW.created_at, NEW.updated_at
        ) ON CONFLICT (payment_id) DO NOTHING;
    END IF;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION orderflow_sync_mock_provider_terminal_status()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.status = 'PENDING' AND NEW.status IN ('SUCCESS', 'FAILED') THEN
        UPDATE mock_provider_transactions
        SET status = NEW.status,
            amount_paise = NEW.amount_paise,
            currency = NEW.currency,
            failure_reason = NEW.failure_reason,
            updated_at = NOW()
        WHERE payment_id = NEW.id AND status = 'PENDING';
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_orderflow_assign_mock_provider_reference ON payments;
CREATE TRIGGER trg_orderflow_assign_mock_provider_reference
BEFORE INSERT ON payments FOR EACH ROW
EXECUTE FUNCTION orderflow_assign_mock_provider_reference();

DROP TRIGGER IF EXISTS trg_orderflow_create_mock_provider_transaction ON payments;
CREATE TRIGGER trg_orderflow_create_mock_provider_transaction
AFTER INSERT ON payments FOR EACH ROW
EXECUTE FUNCTION orderflow_create_mock_provider_transaction();

DROP TRIGGER IF EXISTS trg_orderflow_sync_mock_provider_terminal_status ON payments;
CREATE TRIGGER trg_orderflow_sync_mock_provider_terminal_status
AFTER UPDATE OF status ON payments FOR EACH ROW
EXECUTE FUNCTION orderflow_sync_mock_provider_terminal_status();

CREATE INDEX IF NOT EXISTS idx_mock_provider_status_updated
    ON mock_provider_transactions(status, updated_at ASC);

COMMIT;
