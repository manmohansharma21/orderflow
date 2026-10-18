CREATE TABLE IF NOT EXISTS inventory_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    movement_type TEXT NOT NULL CHECK (movement_type IN (
        'OPENING_BALANCE',
        'RECEIPT',
        'ADJUSTMENT',
        'ORDER_RESERVED',
        'ORDER_RELEASED',
        'ORDER_FULFILLED',
        'REFUND_RESTOCK'
    )),
    on_hand_delta INTEGER NOT NULL DEFAULT 0,
    reserved_delta INTEGER NOT NULL DEFAULT 0,
    on_hand_after INTEGER NOT NULL CHECK (on_hand_after >= 0),
    reserved_after INTEGER NOT NULL CHECK (
        reserved_after >= 0 AND reserved_after <= on_hand_after
    ),
    reason TEXT NOT NULL CHECK (length(btrim(reason)) > 0),
    created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reference_type TEXT,
    reference_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT inventory_movements_nonzero_delta_check
        CHECK (on_hand_delta <> 0 OR reserved_delta <> 0)
);

CREATE INDEX IF NOT EXISTS idx_inventory_movements_product_created
    ON inventory_movements(product_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_inventory_movements_type_created
    ON inventory_movements(movement_type, created_at DESC);

-- Capture the current stock state as an opening balance for existing products.
-- Products at zero stock need no synthetic movement row.
INSERT INTO inventory_movements (
    product_id,
    movement_type,
    on_hand_delta,
    reserved_delta,
    on_hand_after,
    reserved_after,
    reason
)
SELECT
    i.product_id,
    'OPENING_BALANCE',
    i.on_hand_quantity,
    i.reserved_quantity,
    i.on_hand_quantity,
    i.reserved_quantity,
    'Opening balance captured when inventory history was enabled'
FROM inventory AS i
WHERE (i.on_hand_quantity <> 0 OR i.reserved_quantity <> 0)
  AND NOT EXISTS (
      SELECT 1
      FROM inventory_movements AS m
      WHERE m.product_id = i.product_id
        AND m.movement_type = 'OPENING_BALANCE'
  );
