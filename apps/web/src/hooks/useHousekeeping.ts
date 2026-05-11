export interface HousekeepingTask {
  id: string;
  tenant_id: string;
  room_id: string;
  room_number: string;
  floor: string;
  task_type: 'clean' | 'refill' | 'maintenance' | 'inspection';
  priority: number;
  status: 'pending' | 'in_progress' | 'paused' | 'completed' | 'cancelled';
  assigned_to?: string;
  assigned_by?: string;
  shift: string;
  estimated_minutes: number;
  actual_minutes: number;
  started_at?: string;
  completed_at?: string;
  checklist_done?: string[];
  supplies_used?: SupplyUsage[];
  notes?: string;
  damage_report?: any;
  card_mode?: string;
  created_at: string;
  updated_at: string;
}

export interface SupplyUsage {
  item_name: string;
  quantity: number;
}

export interface HousekeepingStaff {
  id: string;
  tenant_id: string;
  user_id?: string;
  name: string;
  role: 'cleaner' | 'inspector' | 'manager';
  active: boolean;
  shift: string;
  floors?: string[];
  max_rooms_per_day: number;
  current_load: number;
  rating: number;
  completed_today: number;
  created_at: string;
}

export interface CleaningChecklistTemplate {
  id: string;
  tenant_id: string;
  room_type: string;
  task_type: string;
  items: ChecklistItem[];
  created_at: string;
}

export interface ChecklistItem {
  step_number: number;
  description: string;
  required: boolean;
}

export interface LockAccessCodeWithMode {
  access_code_id: string;
  code: string;
  assigned_to: string;
  cleaner_mode: 'clean_full' | 'clean_refill' | 'maintenance';
  valid_from?: string;
  valid_until?: string;
  floor?: string;
  room_ids?: string[];
}

export interface SuppliesInventoryItem {
  id: string;
  tenant_id: string;
  item_name: string;
  category: string;
  unit: string;
  current_stock: number;
  reorder_level: number;
  cost_per_unit: number;
  supplier?: string;
  last_restocked?: string;
  created_at: string;
}

export interface StaffLoadItem {
  staff_id: string;
  staff_name: string;
  assigned: number;
  in_progress: number;
  completed: number;
  estimated_minutes: number;
}

export interface FloorProgressItem {
  floor: string;
  total: number;
  dirty: number;
  in_progress: number;
  ready: number;
  percent_done: number;
}

export interface HousekeepingDashboardStats {
  total_rooms: number;
  dirty_rooms: number;
  in_progress_rooms: number;
  ready_rooms: number;
  inspected_rooms: number;
  blocked_rooms: number;
  maintenance_rooms: number;
  unassigned_tasks: number;
  avg_clean_time: number;
  staff_on_duty: number;
  staff_load: StaffLoadItem[];
  floor_progress: FloorProgressItem[];
}

export function useHousekeepingTasks(status?: string, floor?: string, staffId?: string, shift?: string) {
  const params = new URLSearchParams();
  if (status) params.set('status', status);
  if (floor) params.set('floor', floor);
  if (staffId) params.set('staff_id', staffId);
  if (shift) params.set('shift', shift);
  return useApiFetch<HousekeepingTask[]>(`/housekeeping/tasks?${params.toString()}`);
}

export function useHousekeepingDashboard() {
  return useApiFetch<HousekeepingDashboardStats>('/housekeeping/dashboard');
}

export function useHousekeepingStaff(activeOnly = true, shift?: string) {
  const params = new URLSearchParams();
  params.set('active_only', activeOnly.toString());
  if (shift) params.set('shift', shift);
  return useApiFetch<HousekeepingStaff[]>(`/housekeeping/staff?${params.toString()}`);
}

export function useAssignTask() {
  return useApiMutation<void, { task_id: string; staff_id: string; assigned_by?: string }>((body) =>
    apiPost(`/housekeeping/tasks/${body.task_id}/assign`, { staff_id: body.staff_id, assigned_by: body.assigned_by })
  );
}

export function useStartTask() {
  return useApiMutation<void, { task_id: string; card_mode?: string }>((body) =>
    apiPost(`/housekeeping/tasks/${body.task_id}/start`, { card_mode: body.card_mode })
  );
}

export function useCompleteTask() {
  return useApiMutation<void, { task_id: string; actual_minutes: number; checklist_done: string[]; supplies_used: SupplyUsage[]; notes: string }>((body) =>
    apiPost(`/housekeeping/tasks/${body.task_id}/complete`, {
      actual_minutes: body.actual_minutes,
      checklist_done: body.checklist_done,
      supplies_used: body.supplies_used,
      notes: body.notes,
    })
  );
}

export function useBulkAssignTasks() {
  return useApiMutation<{ assigned: number; total: number }, { task_ids: string[]; staff_id: string }>((body) =>
    apiPost('/housekeeping/tasks/bulk-assign', body)
  );
}

export function useSupplies(category?: string) {
  const params = new URLSearchParams();
  if (category) params.set('category', category);
  return useApiFetch<SuppliesInventoryItem[]>(`/housekeeping/supplies?${params.toString()}`);
}

export function useCleanerCards() {
  return useApiFetch<LockAccessCodeWithMode[]>('/housekeeping/cleaner-cards');
}
