-- Housekeeping tasks, staff, checklists, and cleaner card integration
-- Created: 2026-05-11

-- Housekeeping tasks (formalized cleaning assignments)
CREATE TABLE IF NOT EXISTS housekeeping_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    room_id UUID NOT NULL,
    room_number VARCHAR(10) NOT NULL,
    floor VARCHAR(10) NOT NULL,
    task_type VARCHAR(20) NOT NULL DEFAULT 'clean' CHECK (task_type IN ('clean', 'refill', 'maintenance', 'inspection')),
    priority INT NOT NULL DEFAULT 2 CHECK (priority BETWEEN 1 AND 5),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'paused', 'completed', 'cancelled')),
    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    shift VARCHAR(20) NOT NULL DEFAULT 'morning' CHECK (shift IN ('morning', 'afternoon', 'evening', 'night')),
    estimated_minutes INT NOT NULL DEFAULT 30,
    actual_minutes INT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    checklist_done TEXT[],
    supplies_used JSONB DEFAULT '[]',
    notes TEXT,
    damage_report JSONB,
    card_mode VARCHAR(20) CHECK (card_mode IN ('clean_full', 'clean_refill')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_housekeeping_tasks_tenant ON housekeeping_tasks(tenant_id);
CREATE INDEX idx_housekeeping_tasks_status ON housekeeping_tasks(status);
CREATE INDEX idx_housekeeping_tasks_assigned ON housekeeping_tasks(assigned_to);
CREATE INDEX idx_housekeeping_tasks_shift ON housekeeping_tasks(shift);
CREATE INDEX idx_housekeeping_tasks_floor ON housekeeping_tasks(floor);
CREATE INDEX idx_housekeeping_tasks_room ON housekeeping_tasks(room_id);
CREATE INDEX idx_housekeeping_tasks_created ON housekeeping_tasks(created_at DESC);

-- Housekeeping staff profiles
CREATE TABLE IF NOT EXISTS housekeeping_staff (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'cleaner' CHECK (role IN ('cleaner', 'inspector', 'manager')),
    active BOOLEAN NOT NULL DEFAULT true,
    shift VARCHAR(20) NOT NULL DEFAULT 'morning' CHECK (shift IN ('morning', 'afternoon', 'evening', 'night')),
    floors TEXT[],
    max_rooms_per_day INT NOT NULL DEFAULT 15,
    current_load INT NOT NULL DEFAULT 0,
    rating NUMERIC(2,1) NOT NULL DEFAULT 4.0,
    completed_today INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_housekeeping_staff_tenant ON housekeeping_staff(tenant_id);
CREATE INDEX idx_housekeeping_staff_active ON housekeeping_staff(active);
CREATE INDEX idx_housekeeping_staff_shift ON housekeeping_staff(shift);

-- Cleaning checklist templates
CREATE TABLE IF NOT EXISTS cleaning_checklist_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    room_type VARCHAR(50) NOT NULL,
    task_type VARCHAR(20) NOT NULL DEFAULT 'clean' CHECK (task_type IN ('clean', 'refill')),
    items JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_checklist_templates_unique ON cleaning_checklist_templates(tenant_id, room_type, task_type);

-- Staff performance log (daily snapshot)
CREATE TABLE IF NOT EXISTS housekeeping_staff_performance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    staff_id UUID NOT NULL REFERENCES housekeeping_staff(id) ON DELETE CASCADE,
    staff_name VARCHAR(100) NOT NULL,
    performance_date DATE NOT NULL,
    tasks_completed INT NOT NULL DEFAULT 0,
    avg_minutes_per_room NUMERIC(5,2),
    quality_score NUMERIC(2,1),
    complaints INT NOT NULL DEFAULT 0,
    rooms_cleaned INT NOT NULL DEFAULT 0,
    rooms_inspected INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, staff_id, performance_date)
);

CREATE INDEX idx_staff_performance_date ON housekeeping_staff_performance(performance_date DESC);

-- Cleaner card access (smart lock integration)
CREATE TABLE IF NOT EXISTS cleaner_card_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    staff_id UUID NOT NULL REFERENCES housekeeping_staff(id) ON DELETE CASCADE,
    card_code VARCHAR(50) NOT NULL,
    cleaner_mode VARCHAR(20) NOT NULL DEFAULT 'clean_full' CHECK (cleaner_mode IN ('clean_full', 'clean_refill', 'maintenance')),
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMP WITH TIME ZONE,
    floor VARCHAR(10),
    room_ids UUID[] DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cleaner_card_staff ON cleaner_card_access(staff_id);
CREATE INDEX idx_cleaner_card_active ON cleaner_card_access(is_active);

-- Supplies inventory tracking
CREATE TABLE IF NOT EXISTS supplies_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    item_name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL, -- linen, amenity, minibar, cleaning_supply
    unit VARCHAR(20) NOT NULL,     -- piece, roll, bottle, box
    current_stock INT NOT NULL DEFAULT 0,
    reorder_level INT NOT NULL DEFAULT 20,
    cost_per_unit NUMERIC(10,2) NOT NULL DEFAULT 0,
    supplier VARCHAR(100),
    last_restocked TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_supplies_tenant ON supplies_inventory(tenant_id);
CREATE INDEX idx_supplies_category ON supplies_inventory(category);

-- Supplies usage log (linked to tasks)
CREATE TABLE IF NOT EXISTS supplies_usage_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES housekeeping_tasks(id) ON DELETE CASCADE,
    staff_id UUID NOT NULL REFERENCES housekeeping_staff(id) ON DELETE CASCADE,
    item_name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    action VARCHAR(20) NOT NULL DEFAULT 'consumed' CHECK (action IN ('consumed', 'restocked', 'damaged')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_supplies_usage_task ON supplies_usage_log(task_id);
CREATE INDEX idx_supplies_usage_staff ON supplies_usage_log(staff_id);
