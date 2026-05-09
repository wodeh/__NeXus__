-- Rollback: Billing & Subscriptions
-- Version: 000004

DROP TABLE IF EXISTS payment_methods;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS subscription_plans;
