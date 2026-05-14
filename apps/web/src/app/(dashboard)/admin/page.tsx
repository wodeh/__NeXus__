'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth';
import { useTenant } from '@/hooks/useTenant';
import {
  Shield,
  Users,
  Building2,
  Settings,
  BarChart3,
  Key,
  Plus,
  Search,
  MoreHorizontal,
  Edit2,
  Trash2,
  CheckCircle,
  XCircle,
  Lock,
  Globe,
  Clock,
  DollarSign,
  Star,
  ChevronDown,
  ChevronRight,
  AlertTriangle,
  Activity
} from 'lucide-react';
import { getAdminUsers, getAdminProperties, getAdminConfig } from '@/lib/api';

// --- Types ---
interface AdminUser {
  id: string;
  email: string;
  name: string;
  role: string;
  is_active: boolean;
  last_login_at?: string;
  created_at: string;
}

interface AdminProperty {
  id: string;
  name: string;
  address: string;
  city: string;
  country: string;
  phone: string;
  email: string;
  timezone: string;
  currency: string;
  star_rating: number;
  is_active: boolean;
  room_count: number;
  user_count: number;
}

interface SystemConfig {
  default_check_in_time: string;
  default_check_out_time: string;
  auto_confirm: boolean;
  require_deposit: boolean;
  deposit_percent: number;
  allow_walk_in: boolean;
  overbooking_enabled: boolean;
}

// --- Components ---
export default function AdminPage() {
  const router = useRouter();
  const { user } = useAuth();
  const { config } = useTenant();
  const [tab, setTab] = useState<'overview' | 'hotels' | 'users' | 'roles' | 'settings' | 'audit'>('overview');
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [properties, setProperties] = useState<AdminProperty[]>([]);
  const [systemConfig, setSystemConfig] = useState<SystemConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    setLoading(true);
    try {
      const [u, p, c] = await Promise.all([
        getAdminUsers(),
        getAdminProperties(),
        getAdminConfig(),
      ]);
      setUsers(u || []);
      setProperties(p || []);
      setSystemConfig(c);
    } catch (e) {
      console.error('Failed to load admin data', e);
    } finally {
      setLoading(false);
    }
  }

  // Check if user is super_admin
  const isSuperAdmin = user?.role === 'super_admin' || user?.capabilities?.includes('admin:full');

  if (!isSuperAdmin) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-rose-500/10">
          <Lock className="h-8 w-8 text-rose-400" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">Access Denied</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">
          This section is restricted to system administrators only.
        </p>
      </div>
    );
  }

  const filteredUsers = users.filter(u => 
    u.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    u.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
    u.role.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const filteredProperties = properties.filter(p =>
    p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    p.city.toLowerCase().includes(searchQuery.toLowerCase()) ||
    p.country.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Admin Center</h1>
          <p className="text-sm text-slate-400">System administration and hotel management</p>
        </div>
        <div className="flex items-center gap-2">
          <span className="rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-400">
            Super Admin
          </span>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-4">
        <StatCard
          title="Total Hotels"
          value={properties.length.toString()}
          icon={Building2}
          trend="+2 this month"
          color="bg-nexus-500/10 text-nexus-400"
        />
        <StatCard
          title="Total Users"
          value={users.length.toString()}
          icon={Users}
          trend="+5 this week"
          color="bg-sky-500/10 text-sky-400"
        />
        <StatCard
          title="Active Sessions"
          value="12"
          icon={Activity}
          trend="Currently online"
          color="bg-emerald-500/10 text-emerald-400"
        />
        <StatCard
          title="System Health"
          value="99.9%"
          icon={CheckCircle}
          trend="All services operational"
          color="bg-violet-500/10 text-violet-400"
        />
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        {(['overview', 'hotels', 'users', 'roles', 'settings', 'audit'] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              tab === t ? 'bg-slate-800 text-white' : 'text-slate-400 hover:bg-slate-800 hover:text-white'
            }`}
          >
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      {/* Search Bar */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
        <input
          placeholder={`Search ${tab}...`}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="input w-full max-w-md pl-10"
        />
      </div>

      {loading ? (
        <p className="text-sm text-slate-500">Loading admin data...</p>
      ) : (
        <>
          {tab === 'overview' && <OverviewTab properties={properties} users={users} />}
          {tab === 'hotels' && <HotelsTab properties={filteredProperties} />}
          {tab === 'users' && <UsersTab users={filteredUsers} />}
          {tab === 'roles' && <RolesTab />}
          {tab === 'settings' && <SettingsTab config={systemConfig} />}
          {tab === 'audit' && <AuditTab />}
        </>
      )}
    </div>
  );
}

function StatCard({ title, value, icon: Icon, trend, color }: {
  title: string;
  value: string;
  icon: React.ElementType;
  trend: string;
  color: string;
}) {
  return (
    <div className="card space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-sm text-slate-400">{title}</span>
        <div className={`rounded-lg p-2 ${color}`}>
          <Icon className="h-4 w-4" />
        </div>
      </div>
      <p className="text-2xl font-bold text-white">{value}</p>
      <p className="text-xs text-slate-500">{trend}</p>
    </div>
  );
}

function OverviewTab({ properties, users }: { properties: AdminProperty[]; users: AdminUser[] }) {
  const activeUsers = users.filter(u => u.is_active).length;
  const inactiveUsers = users.filter(u => !u.is_active).length;
  const activeProperties = properties.filter(p => p.is_active).length;

  return (
    <div className="grid grid-cols-2 gap-6">
      {/* Recent Hotels */}
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-bold">Recent Hotels</h3>
          <button className="text-xs text-nexus-400 hover:text-nexus-300">View All</button>
        </div>
        <div className="space-y-3">
          {properties.slice(0, 5).map(p => (
            <div key={p.id} className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="flex h-8 w-8 items-center justify-center rounded bg-slate-700 text-sm font-bold">
                  {p.name.charAt(0)}
                </div>
                <div>
                  <p className="text-sm font-medium">{p.name}</p>
                  <p className="text-xs text-slate-500">{p.city}, {p.country}</p>
                </div>
              </div>
              <span className={`rounded-full px-2 py-0.5 text-[10px] ${
                p.is_active ? 'bg-emerald-500/10 text-emerald-400' : 'bg-slate-700 text-slate-500'
              }`}>
                {p.is_active ? 'Active' : 'Inactive'}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* User Distribution */}
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-bold">User Distribution</h3>
          <span className="text-xs text-slate-500">{users.length} total</span>
        </div>
        <div className="space-y-3">
          {['super_admin', 'admin', 'manager', 'front_desk', 'housekeeper', 'readonly'].map(role => {
            const count = users.filter(u => u.role === role).length;
            if (count === 0) return null;
            return (
              <div key={role} className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Users className="h-4 w-4 text-slate-500" />
                  <span className="text-sm capitalize">{role.replace('_', ' ')}</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="h-2 w-24 rounded-full bg-slate-700">
                    <div 
                      className="h-2 rounded-full bg-nexus-500" 
                      style={{ width: `${(count / users.length) * 100}%` }}
                    />
                  </div>
                  <span className="text-xs text-slate-500 w-6">{count}</span>
                </div>
              </div>
            );
          })}
        </div>
        <div className="mt-4 pt-4 border-t border-slate-700 flex justify-between text-xs text-slate-500">
          <span>{activeUsers} active</span>
          <span>{inactiveUsers} inactive</span>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="card col-span-2">
        <h3 className="font-bold mb-4">Quick Actions</h3>
        <div className="grid grid-cols-4 gap-3">
          <QuickActionButton icon={Plus} label="Add Hotel" color="bg-nexus-500/10 text-nexus-400" />
          <QuickActionButton icon={Plus} label="Add User" color="bg-sky-500/10 text-sky-400" />
          <QuickActionButton icon={Key} label="Manage Roles" color="bg-violet-500/10 text-violet-400" />
          <QuickActionButton icon={Settings} label="System Config" color="bg-amber-500/10 text-amber-400" />
        </div>
      </div>
    </div>
  );
}

function QuickActionButton({ icon: Icon, label, color }: { icon: React.ElementType; label: string; color: string }) {
  return (
    <button className="flex items-center gap-3 rounded-lg border border-slate-700 bg-slate-800/50 p-4 hover:bg-slate-700/50 transition-colors">
      <div className={`rounded-lg p-2 ${color}`}>
        <Icon className="h-4 w-4" />
      </div>
      <span className="text-sm font-medium">{label}</span>
    </button>
  );
}

function HotelsTab({ properties }: { properties: AdminProperty[] }) {
  return (
    <div className="space-y-4">
      <div className="flex justify-between">
        <h3 className="text-lg font-bold">Hotel Management</h3>
        <button className="btn-primary gap-2">
          <Plus className="h-4 w-4" />
          Add Hotel
        </button>
      </div>

      <div className="rounded-xl border border-slate-700 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-800/50">
            <tr>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Hotel</th>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Location</th>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Rooms</th>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Users</th>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Rating</th>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Status</th>
              <th className="px-4 py-3 text-right font-medium text-slate-400">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-700">
            {properties.map(p => (
              <tr key={p.id} className="hover:bg-slate-800/30">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-3">
                    <div className="flex h-8 w-8 items-center justify-center rounded bg-slate-700 text-sm font-bold">
                      {p.name.charAt(0)}
                    </div>
                    <div>
                      <p className="font-medium">{p.name}</p>
                      <p className="text-xs text-slate-500">{p.email}</p>
                    </div>
                  </div>
                </td>
                <td className="px-4 py-3 text-slate-400">{p.city}, {p.country}</td>
                <td className="px-4 py-3 text-slate-400">{p.room_count || '-'}</td>
                <td className="px-4 py-3 text-slate-400">{p.user_count || '-'}</td>
                <td className="px-4 py-3">
                  <div className="flex items-center gap-1">
                    <Star className="h-3 w-3 text-amber-400 fill-amber-400" />
                    <span className="text-slate-400">{p.star_rating || '-'}</span>
                  </div>
                </td>
                <td className="px-4 py-3">
                  <span className={`rounded-full px-2 py-0.5 text-xs ${
                    p.is_active ? 'bg-emerald-500/10 text-emerald-400' : 'bg-slate-700 text-slate-500'
                  }`}>
                    {p.is_active ? 'Active' : 'Inactive'}
                  </span>
                </td>
                <td className="px-4 py-3 text-right">
                  <div className="flex items-center justify-end gap-2">
                    <button className="rounded p-1 text-slate-400 hover:text-white">
                      <Edit2 className="h-3 w-3" />
                    </button>
                    <button className="rounded p-1 text-slate-400 hover:text-rose-400">
                      <Trash2 className="h-3 w-3" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function UsersTab({ users }: { users: AdminUser[] }) {
  return (
    <div className="space-y-4">
      <div className="flex justify-between">
        <h3 className="text-lg font-bold">User Management</h3>
        <button className="btn-primary gap-2">
          <Plus className="h-4 w-4" />
          Add User
        </button>
      </div>

      <div className="rounded-xl border border-slate-700 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-800/50">
            <tr>
              <th className="px-4 py-3 text-left font-medium text-slate-400">User</th>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Role</th>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Status</th>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Last Login</th>
              <th className="px-4 py-3 text-right font-medium text-slate-400">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-700">
            {users.map(u => (
              <tr key={u.id} className="hover:bg-slate-800/30">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-3">
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-slate-700 text-sm font-bold">
                      {u.name.charAt(0)}
                    </div>
                    <div>
                      <p className="font-medium">{u.name}</p>
                      <p className="text-xs text-slate-500">{u.email}</p>
                    </div>
                  </div>
                </td>
                <td className="px-4 py-3">
                  <span className="rounded-full bg-slate-700 px-2 py-0.5 text-xs capitalize">
                    {u.role.replace('_', ' ')}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className={`flex items-center gap-1 text-xs ${
                    u.is_active ? 'text-emerald-400' : 'text-slate-500'
                  }`}>
                    {u.is_active ? <CheckCircle className="h-3 w-3" /> : <XCircle className="h-3 w-3" />}
                    {u.is_active ? 'Active' : 'Inactive'}
                  </span>
                </td>
                <td className="px-4 py-3 text-slate-500 text-xs">
                  {u.last_login_at ? new Date(u.last_login_at).toLocaleDateString() : 'Never'}
                </td>
                <td className="px-4 py-3 text-right">
                  <div className="flex items-center justify-end gap-2">
                    <button className="rounded p-1 text-slate-400 hover:text-white">
                      <Edit2 className="h-3 w-3" />
                    </button>
                    <button className="rounded p-1 text-slate-400 hover:text-rose-400">
                      <Trash2 className="h-3 w-3" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function RolesTab() {
  const roles = [
    { name: 'super_admin', description: 'Full platform access', permissions: ['*'], userCount: 1 },
    { name: 'admin', description: 'Hotel administrator', permissions: ['reservations:read', 'reservations:write', 'rooms:read', 'rooms:write', 'users:read', 'users:write'], userCount: 2 },
    { name: 'manager', description: 'Operations manager', permissions: ['reservations:read', 'reservations:write', 'housekeeping:read', 'reports:read'], userCount: 3 },
    { name: 'front_desk', description: 'Front desk agent', permissions: ['reservations:read', 'reservations:write', 'guests:read'], userCount: 5 },
    { name: 'housekeeper', description: 'Housekeeping staff', permissions: ['housekeeping:read', 'housekeeping:write', 'rooms:read'], userCount: 4 },
    { name: 'readonly', description: 'Read-only access', permissions: ['reservations:read', 'rooms:read', 'housekeeping:read'], userCount: 0 },
  ];

  return (
    <div className="space-y-4">
      <div className="flex justify-between">
        <h3 className="text-lg font-bold">Role Management</h3>
        <button className="btn-primary gap-2">
          <Plus className="h-4 w-4" />
          Create Role
        </button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        {roles.map(role => (
          <div key={role.name} className="card space-y-3">
            <div className="flex items-center justify-between">
              <h4 className="font-bold capitalize">{role.name.replace('_', ' ')}</h4>
              <span className="text-xs text-slate-500">{role.userCount} users</span>
            </div>
            <p className="text-sm text-slate-400">{role.description}</p>
            <div className="flex flex-wrap gap-1">
              {role.permissions.map(p => (
                <span key={p} className="rounded bg-slate-700 px-2 py-0.5 text-[10px] text-slate-400">
                  {p}
                </span>
              ))}
            </div>
            <div className="flex gap-2 pt-2">
              <button className="text-xs text-nexus-400 hover:text-nexus-300">Edit</button>
              <button className="text-xs text-slate-500 hover:text-slate-400">Clone</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function SettingsTab({ config }: { config: SystemConfig | null }) {
  if (!config) return <p className="text-slate-500">Loading configuration...</p>;

  return (
    <div className="max-w-2xl space-y-6">
      <h3 className="text-lg font-bold">System Configuration</h3>

      <div className="card space-y-4">
        <h4 className="font-medium text-slate-400">Reservation Defaults</h4>
        
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs text-slate-500 mb-1">Default Check-in Time</label>
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-slate-500" />
              <input className="input w-full" defaultValue={config.default_check_in_time} />
            </div>
          </div>
          <div>
            <label className="block text-xs text-slate-500 mb-1">Default Check-out Time</label>
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-slate-500" />
              <input className="input w-full" defaultValue={config.default_check_out_time} />
            </div>
          </div>
        </div>

        <div className="space-y-3">
          <ToggleSetting label="Auto-confirm reservations" defaultChecked={config.auto_confirm} />
          <ToggleSetting label="Require deposit" defaultChecked={config.require_deposit} />
          <ToggleSetting label="Allow walk-ins" defaultChecked={config.allow_walk_in} />
          <ToggleSetting label="Enable overbooking" defaultChecked={config.overbooking_enabled} />
        </div>

        <div>
          <label className="block text-xs text-slate-500 mb-1">Deposit Percentage</label>
          <div className="flex items-center gap-2">
            <DollarSign className="h-4 w-4 text-slate-500" />
            <input 
              type="number" 
              className="input w-24" 
              defaultValue={config.deposit_percent} 
              min="0" 
              max="100" 
            />
            <span className="text-sm text-slate-500">%</span>
          </div>
        </div>
      </div>

      <div className="card space-y-4">
        <h4 className="font-medium text-slate-400">Regional Settings</h4>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs text-slate-500 mb-1">Default Timezone</label>
            <div className="flex items-center gap-2">
              <Globe className="h-4 w-4 text-slate-500" />
              <select className="input w-full">
                <option>UTC</option>
                <option>America/New_York</option>
                <option>Europe/London</option>
                <option>Asia/Dubai</option>
              </select>
            </div>
          </div>
          <div>
            <label className="block text-xs text-slate-500 mb-1">Default Currency</label>
            <div className="flex items-center gap-2">
              <DollarSign className="h-4 w-4 text-slate-500" />
              <select className="input w-full">
                <option>USD</option>
                <option>EUR</option>
                <option>GBP</option>
                <option>AED</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      <button className="btn-primary w-full">Save Configuration</button>
    </div>
  );
}

function ToggleSetting({ label, defaultChecked }: { label: string; defaultChecked: boolean }) {
  const [checked, setChecked] = useState(defaultChecked);
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm">{label}</span>
      <button 
        onClick={() => setChecked(!checked)}
        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
          checked ? 'bg-nexus-500' : 'bg-slate-700'
        }`}
      >
        <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
          checked ? 'translate-x-6' : 'translate-x-1'
        }`} />
      </button>
    </div>
  );
}

function AuditTab() {
  const logs = [
    { action: 'User created', user: 'admin@system.com', target: 'john@hotel.com', time: '2026-05-14 10:23', type: 'info' },
    { action: 'Property updated', user: 'admin@system.com', target: 'Grand Plaza', time: '2026-05-14 09:15', type: 'warning' },
    { action: 'Role modified', user: 'super_admin', target: 'front_desk', time: '2026-05-13 16:45', type: 'info' },
    { action: 'Login failed', user: 'unknown', target: '192.168.1.100', time: '2026-05-13 14:20', type: 'error' },
    { action: 'Settings changed', user: 'admin@system.com', target: 'System Config', time: '2026-05-13 11:00', type: 'warning' },
  ];

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-bold">Audit Log</h3>
      
      <div className="rounded-xl border border-slate-700 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-800/50">
            <tr>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Action</th>
              <th className="px-4 py-3 text-left font-medium text-slate-400">User</th>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Target</th>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Time</th>
              <th className="px-4 py-3 text-left font-medium text-slate-400">Severity</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-700">
            {logs.map((log, i) => (
              <tr key={i} className="hover:bg-slate-800/30">
                <td className="px-4 py-3 font-medium">{log.action}</td>
                <td className="px-4 py-3 text-slate-400">{log.user}</td>
                <td className="px-4 py-3 text-slate-400">{log.target}</td>
                <td className="px-4 py-3 text-slate-500 text-xs">{log.time}</td>
                <td className="px-4 py-3">
                  <span className={`rounded-full px-2 py-0.5 text-xs ${
                    log.type === 'error' ? 'bg-rose-500/10 text-rose-400' :
                    log.type === 'warning' ? 'bg-amber-500/10 text-amber-400' :
                    'bg-sky-500/10 text-sky-400'
                  }`}>
                    {log.type}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
