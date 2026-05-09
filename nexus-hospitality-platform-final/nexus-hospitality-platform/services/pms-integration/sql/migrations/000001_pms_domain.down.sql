-- Rollback PMS Domain Foundation Migration

DROP TRIGGER IF EXISTS update_folio_transactions_updated_at ON folio_transactions;
DROP TRIGGER IF EXISTS update_folios_updated_at ON folios;
DROP TRIGGER IF EXISTS update_reservations_updated_at ON reservations;
DROP TRIGGER IF EXISTS update_guests_updated_at ON guests;

DROP TABLE IF EXISTS folio_transactions;
DROP TABLE IF EXISTS folios;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS guests;
