// clients/web-portal/src/App.tsx
// ============================================================
// NEXUS WEB PORTAL — Micro-Frontend Orchestrator
// ============================================================

import React, { Suspense, lazy, useEffect, useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Provider as JotaiProvider } from 'jotai';
import { I18nextProvider } from 'react-i18next';
import { ErrorBoundary } from 'react-error-boundary';
import { Toaster } from 'sonner';

import { ThemeProvider } from './hooks/useTheme';
import { WebSocketProvider } from './hooks/useWebSocket';
import { AuthProvider } from './hooks/useAuth';
import { RBACProvider } from './hooks/useRBAC';
import { OfflineProvider } from './hooks/useOffline';
import { TelemetryProvider } from './hooks/useTelemetry';

import { MainLayout } from './layouts/MainLayout';
import { LoadingScreen } from './components/LoadingScreen';
import { ErrorFallback } from './components/ErrorFallback';

import i18n from './i18n';

// Lazy-loaded micro-frontends (module federation)
const GuestPortal = lazy(() => import('guestPortal/GuestPortal'));
const StaffConsole = lazy(() => import('staffConsole/StaffConsole'));
const ExecutiveDashboard = lazy(() => import('executiveDashboard/ExecutiveDashboard'));
const FranchisePortal = lazy(() => import('franchisePortal/FranchisePortal'));
const MSPPortal = lazy(() => import('mspPortal/MSPPortal'));
const IPTVManager = lazy(() => import('iptvManager/IPTVManager'));
const IoTControlPanel = lazy(() => import('iotControlPanel/IoTControlPanel'));
const AIAnalytics = lazy(() => import('aiAnalytics/AIAnalytics'));

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      cacheTime: 30 * 60 * 1000, // 30 minutes
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      refetchOnWindowFocus: false,
      suspense: true,
    },
    mutations: {
      retry: 1,
    },
  },
});

export const App: React.FC = () => {
  const [isOnline, setIsOnline] = useState(navigator.onLine);

  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  return (
    <ErrorBoundary FallbackComponent={ErrorFallback}>
      <JotaiProvider>
        <QueryClientProvider client={queryClient}>
          <I18nextProvider i18n={i18n}>
            <ThemeProvider>
              <AuthProvider>
                <RBACProvider>
                  <WebSocketProvider>
                    <OfflineProvider isOnline={isOnline}>
                      <TelemetryProvider>
                        <BrowserRouter>
                          <Routes>
                            <Route path="/" element={<MainLayout />}>
                              {/* Guest-facing routes */}
                              <Route
                                path="guest/*"
                                element={
                                  <Suspense fallback={<LoadingScreen />}>
                                    <GuestPortal />
                                  </Suspense>
                                }
                              />

                              {/* Staff operations */}
                              <Route
                                path="staff/*"
                                element={
                                  <Suspense fallback={<LoadingScreen />}>
                                    <StaffConsole />
                                  </Suspense>
                                }
                              />

                              {/* Executive/Analytics */}
                              <Route
                                path="executive/*"
                                element={
                                  <Suspense fallback={<LoadingScreen />}>
                                    <ExecutiveDashboard />
                                  </Suspense>
                                }
                              />

                              {/* Franchise management */}
                              <Route
                                path="franchise/*"
                                element={
                                  <Suspense fallback={<LoadingScreen />}>
                                    <FranchisePortal />
                                  </Suspense>
                                }
                              />

                              {/* MSP/Tech partner */}
                              <Route
                                path="msp/*"
                                element={
                                  <Suspense fallback={<LoadingScreen />}>
                                    <MSPPortal />
                                  </Suspense>
                                }
                              />

                              {/* IPTV management */}
                              <Route
                                path="iptv/*"
                                element={
                                  <Suspense fallback={<LoadingScreen />}>
                                    <IPTVManager />
                                  </Suspense>
                                }
                              />

                              {/* IoT/Smart Room */}
                              <Route
                                path="iot/*"
                                element={
                                  <Suspense fallback={<LoadingScreen />}>
                                    <IoTControlPanel />
                                  </Suspense>
                                }
                              />

                              {/* AI/Analytics */}
                              <Route
                                path="ai/*"
                                element={
                                  <Suspense fallback={<LoadingScreen />}>
                                    <AIAnalytics />
                                  </Suspense>
                                }
                              />

                              {/* Default redirect based on role */}
                              <Route path="*" element={<RoleBasedRedirect />} />
                            </Route>
                          </Routes>
                        </BrowserRouter>
                        <Toaster
                          position="top-right"
                          richColors
                          closeButton
                          toastOptions={{
                            style: {
                              fontFamily: 'Inter, sans-serif',
                            },
                          }}
                        />
                      </TelemetryProvider>
                    </OfflineProvider>
                  </WebSocketProvider>
                </RBACProvider>
              </AuthProvider>
            </ThemeProvider>
          </I18nextProvider>
        </QueryClientProvider>
      </JotaiProvider>
    </ErrorBoundary>
  );
};

const RoleBasedRedirect: React.FC = () => {
  const { user } = useAuth();

  if (!user) return <Navigate to="/guest" replace />;

  switch (user.role) {
    case 'guest':
      return <Navigate to="/guest" replace />;
    case 'housekeeping':
    case 'front_desk':
    case 'manager':
      return <Navigate to="/staff" replace />;
    case 'executive':
      return <Navigate to="/executive" replace />;
    case 'franchise_owner':
      return <Navigate to="/franchise" replace />;
    case 'msp_engineer':
      return <Navigate to="/msp" replace />;
    default:
      return <Navigate to="/guest" replace />;
  }
};

// RBAC-aware component wrapper
export const RBACGuard: React.FC<{
  requiredPermissions: string[];
  requiredRoles?: string[];
  fallback?: React.ReactNode;
  children: React.ReactNode;
}> = ({ requiredPermissions, requiredRoles, fallback, children }) => {
  const { hasPermission, hasRole } = useRBAC();

  const authorized = 
    requiredPermissions.every(p => hasPermission(p)) &&
    (!requiredRoles || requiredRoles.some(r => hasRole(r)));

  if (!authorized) {
    return fallback || (
      <div className="p-8 text-center">
        <h2 className="text-xl font-semibold text-red-600">Access Denied</h2>
        <p className="text-gray-600 mt-2">
          You don't have permission to access this feature.
        </p>
      </div>
    );
  }

  return <>{children}</>;
};

// Real-time event synchronization hook
export const useRealTimeSync = (channel: string) => {
  const { socket } = useWebSocket();
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!socket) return;

    const handleEvent = (event: any) => {
      // Invalidate relevant queries based on event type
      switch (event.type) {
        case 'reservation_updated':
          queryClient.invalidateQueries(['reservations']);
          break;
        case 'room_status_changed':
          queryClient.invalidateQueries(['rooms']);
          break;
        case 'housekeeping_task_completed':
          queryClient.invalidateQueries(['housekeeping']);
          break;
        case 'maintenance_request_created':
          queryClient.invalidateQueries(['maintenance']);
          break;
        case 'guest_checked_in':
          queryClient.invalidateQueries(['guests', 'reservations']);
          break;
        case 'iptv_stream_started':
          queryClient.invalidateQueries(['iptv', 'analytics']);
          break;
        case 'iot_device_telemetry':
          queryClient.invalidateQueries(['iot', 'telemetry']);
          break;
        default:
          // Generic invalidation for unknown events
          queryClient.invalidateQueries();
      }
    };

    socket.on(channel, handleEvent);
    return () => {
      socket.off(channel, handleEvent);
    };
  }, [socket, channel, queryClient]);
};

export default App;
