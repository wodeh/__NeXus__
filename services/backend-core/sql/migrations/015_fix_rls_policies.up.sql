-- Fix broken RLS policies
-- Previous policies used current_setting() which returns NULL when unset,
-- causing all rows to be invisible (NULL OR NULL → NULL, not TRUE).
-- Drop policies until we implement proper per-connection tenant binding.

DROP POLICY IF EXISTS rooms_isolation_policy ON rooms;
DROP POLICY IF EXISTS reservations_isolation_policy ON reservations;
DROP POLICY IF EXISTS hk_staff_isolation_policy ON housekeeping_staff;
DROP POLICY IF EXISTS hk_tasks_isolation_policy ON housekeeping_tasks;

-- Re-enable RLS without policies (superuser/owner bypass)
-- When we re-enable, we'll use proper connection-scoped tenant binding
