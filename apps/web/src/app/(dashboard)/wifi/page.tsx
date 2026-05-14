'use client';

import { useState, useEffect, useMemo } from 'react';
import {
  Wifi, WifiOff, Signal, AlertTriangle, CheckCircle2, RefreshCw, Activity,
  MapPin, ChevronRight, Zap, Radio, Eye, BarChart3, Gauge, Settings
} from 'lucide-react';
import { getWiFiDashboard, getWiFiHeatmap, resolveWiFiAlert, triggerWiFiScan, WiFiAlert, WiFiFloorSummary } from '@/lib/api';

export default function WiFiMonitoringPage() {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedFloor, setSelectedFloor] = useState<string>('');
  const [heatmapData, setHeatmapData] = useState<any>(null);
  const [scanning, setScanning] = useState(false);
  const [activeTab, setActiveTab] = useState<'overview' | 'heatmap' | 'alerts' | 'aps'>('overview');

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      const dash = await getWiFiDashboard();
      setData(dash);
      if (dash.floors.length > 0 && !selectedFloor) {
        setSelectedFloor(dash.floors[0].floor);
      }
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  useEffect(() => {
    if (!selectedFloor) return;
    getWiFiHeatmap(selectedFloor).then(setHeatmapData).catch(() => {});
  }, [selectedFloor]);

  const handleResolveAlert = async (alertId: string) => {
    await resolveWiFiAlert(alertId);
    fetchData();
  };

  const handleScan = async () => {
    setScanning(true);
    await triggerWiFiScan(selectedFloor);
    setTimeout(() => { setScanning(false); fetchData(); }, 3000);
  };

  if (loading) {
    return (
      <div className="flex h-96 items-center justify-center">
        <RefreshCw className="h-8 w-8 animate-spin text-nexus-400" />
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="flex h-96 flex-col items-center justify-center gap-4">
        <WifiOff className="h-10 w-10 text-rose-400" />
        <p className="text-rose-400">{error || 'No data available'}</p>
        <button onClick={fetchData} className="btn-primary">Retry</button>
      </div>
    );
  }

  const overview = data.overview;
  const floors: WiFiFloorSummary[] = data.floors || [];
  const alerts: WiFiAlert[] = data.alerts || [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">WiFi Network Monitor</h1>
          <p className="text-sm text-slate-400">Enterprise AP monitoring via SNMP</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={handleScan} disabled={scanning} className="btn-primary">
            {scanning ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Activity className="h-4 w-4" />}
            {scanning ? 'Scanning...' : 'SNMP Scan'}
          </button>
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white">
            <RefreshCw className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* Overview Stats */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-5">
        <StatCard label="Total APs" value={overview.total_aps} sub={`${overview.online_aps} online`} icon={Radio} color="sky" />
        <StatCard label="Clients" value={overview.total_clients} sub="connected devices" icon={Wifi} color="emerald" />
        <StatCard label="Active Alerts" value={overview.active_alerts} sub="needs attention" icon={AlertTriangle} color={overview.active_alerts > 0 ? 'rose' : 'emerald'} />
        <StatCard label="Avg Signal" value={`${overview.avg_signal_dbm} dBm`} sub={overview.overall_health} icon={Signal} color={signalColor(overview.avg_signal_dbm)} />
        <StatCard label="Network Health" value={overview.overall_health} sub="overall status" icon={CheckCircle2} color={healthColor(overview.overall_health)} />
      </div>

      {/* Tabs */}
      <div className="flex gap-1 rounded-lg border border-slate-700 bg-slate-800 p-1">
        {(['overview', 'heatmap', 'alerts', 'aps'] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`flex-1 rounded-md px-4 py-2 text-sm font-medium transition-colors ${
              activeTab === tab ? 'bg-nexus-500/20 text-nexus-400' : 'text-slate-400 hover:text-white'
            }`}
          >
            {tab === 'overview' && <BarChart3 className="mr-2 inline h-4 w-4" />}
            {tab === 'heatmap' && <MapPin className="mr-2 inline h-4 w-4" />}
            {tab === 'alerts' && <AlertTriangle className="mr-2 inline h-4 w-4" />}
            {tab === 'aps' && <Radio className="mr-2 inline h-4 w-4" />}
            {tab.charAt(0).toUpperCase() + tab.slice(1)}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      {activeTab === 'overview' && (
        <div className="grid gap-6 lg:grid-cols-3">
          {/* Floor Cards */}
          <div className="lg:col-span-2 space-y-4">
            <h3 className="font-bold text-white">Floor Status</h3>
            <div className="grid gap-4 sm:grid-cols-2">
              {floors.map((floor) => (
                <FloorCard
                  key={floor.floor}
                  floor={floor}
                  isSelected={selectedFloor === floor.floor}
                  onClick={() => setSelectedFloor(floor.floor)}
                />
              ))}
            </div>
          </div>

          {/* Alerts Panel */}
          <div className="space-y-4">
            <h3 className="font-bold text-white">Active Alerts</h3>
            <div className="space-y-3">
              {alerts.length === 0 && (
                <div className="rounded-lg border border-slate-700 bg-slate-800 p-4 text-center">
                  <CheckCircle2 className="mx-auto h-8 w-8 text-emerald-400" />
                  <p className="mt-2 text-sm text-slate-400">No active alerts</p>
                </div>
              )}
              {alerts.map((alert) => (
                <AlertCard key={alert.id} alert={alert} onResolve={() => handleResolveAlert(alert.id)} />
              ))}
            </div>
          </div>
        </div>
      )}

      {activeTab === 'heatmap' && selectedFloor && heatmapData && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-bold text-white">Signal Heatmap — {selectedFloor}</h3>
            <select
              value={selectedFloor}
              onChange={(e) => setSelectedFloor(e.target.value)}
              className="rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white"
            >
              {floors.map((f) => (
                <option key={f.floor} value={f.floor}>{f.floor}</option>
              ))}
            </select>
          </div>
          <HeatmapGrid data={heatmapData} />
          <div className="flex gap-4 text-xs text-slate-400">
            <span className="flex items-center gap-1"><span className="h-3 w-3 rounded bg-emerald-500" /> Excellent (-30 to -50 dBm)</span>
            <span className="flex items-center gap-1"><span className="h-3 w-3 rounded bg-sky-500" /> Good (-50 to -60 dBm)</span>
            <span className="flex items-center gap-1"><span className="h-3 w-3 rounded bg-amber-500" /> Fair (-60 to -70 dBm)</span>
            <span className="flex items-center gap-1"><span className="h-3 w-3 rounded bg-rose-500" /> Poor (-70+ dBm)</span>
          </div>
        </div>
      )}

      {activeTab === 'alerts' && (
        <div className="space-y-4">
          <h3 className="font-bold text-white">All Alerts</h3>
          <div className="space-y-3">
            {alerts.map((alert) => (
              <AlertCard key={alert.id} alert={alert} onResolve={() => handleResolveAlert(alert.id)} />
            ))}
          </div>
        </div>
      )}

      {activeTab === 'aps' && (
        <div className="space-y-4">
          <h3 className="font-bold text-white">Access Points</h3>
          <div className="overflow-x-auto rounded-lg border border-slate-700">
            <table className="w-full text-sm">
              <thead className="bg-slate-800">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-slate-400">Name</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-slate-400">Floor</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-slate-400">MAC</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-slate-400">Status</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-slate-400">Model</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-slate-400">Clients</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700">
                {/* Demo AP rows */}
                {[
                  { name: 'Lobby-AP-01', floor: 'Lobby', mac: '00:11:22:33:44:AA', status: 'online', model: 'Ubiquiti U6-Pro', clients: 23 },
                  { name: 'Lobby-AP-02', floor: 'Lobby', mac: '00:11:22:33:44:AB', status: 'online', model: 'Ubiquiti U6-Pro', clients: 19 },
                  { name: 'F1-AP-01', floor: 'Floor 1', mac: '00:11:22:33:44:A1', status: 'online', model: 'Ubiquiti U6-Lite', clients: 12 },
                  { name: 'F2-AP-01', floor: 'Floor 2', mac: '00:11:22:33:44:B1', status: 'offline', model: 'Ubiquiti U6-Lite', clients: 0 },
                  { name: 'F5-AP-03', floor: 'Floor 5', mac: '00:11:22:33:44:E3', status: 'degraded', model: 'Ubiquiti U6-Lite', clients: 6 },
                ].map((ap) => (
                  <tr key={ap.mac} className="hover:bg-slate-700/30">
                    <td className="px-4 py-3 text-white">{ap.name}</td>
                    <td className="px-4 py-3 text-slate-400">{ap.floor}</td>
                    <td className="px-4 py-3 text-slate-400">{ap.mac}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs ${
                        ap.status === 'online' ? 'bg-emerald-500/10 text-emerald-400' :
                        ap.status === 'offline' ? 'bg-rose-500/10 text-rose-400' :
                        'bg-amber-500/10 text-amber-400'
                      }`}>
                        {ap.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-400">{ap.model}</td>
                    <td className="px-4 py-3 text-slate-400">{ap.clients}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}

/* ─── Sub-components ─── */

function FloorCard({ floor, isSelected, onClick }: { floor: WiFiFloorSummary; isSelected: boolean; onClick: () => void }) {
  const quality = signalQuality(floor.avg_signal_dbm || -60);
  return (
    <div
      onClick={onClick}
      className={`cursor-pointer rounded-lg border p-4 transition-all ${
        isSelected ? 'border-nexus-500 bg-nexus-500/5' : 'border-slate-700 bg-slate-800 hover:border-slate-600'
      }`}
    >
      <div className="flex items-center justify-between mb-3">
        <h4 className="font-bold text-white">{floor.floor}</h4>
        <span className={`text-[10px] rounded-full px-2 py-0.5 ${quality.color.replace('text-', 'bg-').replace('400', '500/10')} ${quality.color}`}>
          {quality.label}
        </span>
      </div>
      <div className="grid grid-cols-3 gap-2 text-center">
        <div>
          <p className="text-lg font-bold text-white">{floor.online_aps}/{floor.ap_count}</p>
          <p className="text-[10px] text-slate-500">APs Online</p>
        </div>
        <div>
          <p className="text-lg font-bold text-white">{floor.total_clients}</p>
          <p className="text-[10px] text-slate-500">Clients</p>
        </div>
        <div>
          <p className="text-lg font-bold text-white">{floor.active_alerts}</p>
          <p className="text-[10px] text-slate-500">Alerts</p>
        </div>
      </div>
      {/* Signal bars */}
      <div className="mt-3 flex items-center gap-1">
        <span className="text-[10px] text-slate-500 mr-1">Signal</span>
        {[1, 2, 3, 4].map((bar) => (
          <div
            key={bar}
            className={`w-1.5 rounded-sm ${bar <= quality.bars ? quality.color.replace('text-', 'bg-') : 'bg-slate-700'}`}
            style={{ height: `${bar * 3}px` }}
          />
        ))}
        <span className="text-[10px] text-slate-400 ml-2">{floor.avg_signal_dbm} dBm</span>
      </div>
    </div>
  );
}

function AlertCard({ alert, onResolve }: { alert: WiFiAlert; onResolve: () => void }) {
  const colors = {
    critical: 'border-rose-500/30 bg-rose-500/5',
    warning: 'border-amber-500/30 bg-amber-500/5',
    info: 'border-sky-500/30 bg-sky-500/5',
  };
  const iconColors = {
    critical: 'text-rose-400',
    warning: 'text-amber-400',
    info: 'text-sky-400',
  };

  return (
    <div className={`rounded-lg border p-3 ${colors[alert.severity]}`}>
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-2">
          <AlertTriangle className={`h-4 w-4 mt-0.5 ${iconColors[alert.severity]}`} />
          <div>
            <p className="text-sm font-medium text-white">{alert.message}</p>
            <p className="text-xs text-slate-400">{alert.floor}</p>
            {alert.suggested_fix && (
              <p className="text-xs text-slate-500 mt-1">Fix: {alert.suggested_fix}</p>
            )}
          </div>
        </div>
        <button
          onClick={onResolve}
          className="rounded bg-slate-700 px-2 py-1 text-[10px] text-white hover:bg-slate-600"
        >
          Resolve
        </button>
      </div>
    </div>
  );
}

function HeatmapGrid({ data }: { data: any }) {
  const grid: number[][] = data.grid;
  if (!grid) return null;

  return (
    <div className="rounded-lg border border-slate-700 bg-slate-800 p-4 overflow-x-auto">
      <div className="inline-block">
        {grid.map((row: number[], y: number) => (
          <div key={y} className="flex">
            {row.map((cell: number, x: number) => {
              const color = cell >= -50 ? 'bg-emerald-500' :
                           cell >= -60 ? 'bg-sky-500' :
                           cell >= -70 ? 'bg-amber-500' : 'bg-rose-500';
              return (
                <div
                  key={`${x}-${y}`}
                  className={`h-8 w-8 ${color} opacity-80 hover:opacity-100 transition-opacity`}
                  title={`${cell} dBm`}
                />
              );
            })}
          </div>
        ))}
      </div>
    </div>
  );
}

function StatCard({ label, value, sub, icon: Icon, color }: { label: string; value: string | number; sub: string; icon: any; color: string }) {
  const colorMap: Record<string, string> = {
    sky: 'bg-sky-500/10 text-sky-400',
    emerald: 'bg-emerald-500/10 text-emerald-400',
    violet: 'bg-violet-500/10 text-violet-400',
    amber: 'bg-amber-500/10 text-amber-400',
    rose: 'bg-rose-500/10 text-rose-400',
  };
  return (
    <div className="rounded-lg border border-slate-700 bg-slate-800 p-4">
      <div className={`rounded-md p-1.5 w-fit ${colorMap[color] || ''}`}>
        <Icon className="h-5 w-5" />
      </div>
      <div className="mt-2">
        <div className="text-2xl font-bold text-white">{value}</div>
        <div className="text-xs text-slate-400">{sub}</div>
      </div>
    </div>
  );
}

/* ─── Helpers ─── */
function signalQuality(dbm: number): { label: string; color: string; bars: number } {
  if (dbm >= -50) return { label: 'Excellent', color: 'text-emerald-400', bars: 4 };
  if (dbm >= -60) return { label: 'Good', color: 'text-sky-400', bars: 3 };
  if (dbm >= -70) return { label: 'Fair', color: 'text-amber-400', bars: 2 };
  return { label: 'Poor', color: 'text-rose-400', bars: 1 };
}

function signalColor(dbm: number): string {
  if (dbm >= -50) return 'emerald';
  if (dbm >= -60) return 'sky';
  if (dbm >= -70) return 'amber';
  return 'rose';
}

function healthColor(health: string): string {
  if (health === 'excellent') return 'emerald';
  if (health === 'good') return 'sky';
  if (health === 'fair') return 'amber';
  return 'rose';
}
