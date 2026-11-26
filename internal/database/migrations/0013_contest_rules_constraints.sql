-- Aturan contest: kunci nilai severity sesuai kontrak klien dan percepat query per event.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'contest_rules_severity_check'
          AND conrelid = 'contest_rules'::regclass
    ) THEN
        ALTER TABLE contest_rules
            ADD CONSTRAINT contest_rules_severity_check
            CHECK (severity IN ('info', 'error', 'success'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS contest_rules_event_id_sort_order_idx
    ON contest_rules (event_id, sort_order, id);
