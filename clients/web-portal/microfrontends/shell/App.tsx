// clients/web-portal/microfrontends/shell/App.tsx
import React, { Suspense, lazy } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ErrorBoundary } from 'react-error-boundary';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider } from './contexts/AuthContext';
import { TenantProvider } from './contexts/TenantContext';
import { RBACProvider } from './contexts/RBACContext';
import { WebSocketProvider } from './contexts/WebSocketContext';
import { LocalizationProvider } from './contexts/LocalizationContext';
import { OfflineProvider } from './contexts/OfflineContext';
import { DesignSystemProvider } from '@nexus/design-system';
import { LoadingScreen } from '@nexus/design-system/components';
import { Layout } from './components/Layout';
import { GlobalErrorFallback } from './components/GlobalErrorFallback';

// Lazy-loaded microfrontends
const PMSMicrofrontend = lazy(() => import('pms_mf/PMSApp'));
const IPTVMicrofrontend = lazy(() => import('iptv_mf/IPTVApp'));
const IoTMicrofrontend = lazy(() => import('iot_mf/IoTApp'));
const AIMicrofrontend = lazy(() => import('ai_mf/AIApp'));
const BillingMicrofrontend = lazy(() => import('billing_mf/BillingApp'));
const AnalyticsMicrofrontend = lazy(() => import('analytics_mf/AnalyticsApp'));
const SettingsMicrofrontend = lazy(() => import('settings_mf/SettingsApp'));

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      cacheTime: 10 * 60 * 1000,
      retry: 3,
      retryDelay: attemptIndex => Math.min(1000 * 2 ** attemptIndex, 30000),
      refetchOnWindowFocus: false,
      networkMode: 'offlineFirst',
    },
  },
});

function App() {
  return (
    <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <AuthProvider>
            <TenantProvider>
              <RBACProvider>
                <WebSocketProvider>
                  <LocalizationProvider>
                    <OfflineProvider>
                      <DesignSystemProvider theme="enterprise">
                        <Layout>
                          <Suspense fallback={<LoadingScreen fullScreen />}>
                            <Routes>
                              <Route path="/" element={<Navigate to="/pms" replace />} />
                              <Route path="/pms/*" element={<PMSMicrofrontend />} />
                              <Route path="/iptv/*" element={<IPTVMicrofrontend />} />
                              <Route path="/iot/*" element={<IoTMicrofrontend />} />
                              <Route path="/ai/*" element={<AIMicrofrontend />} />
                              <Route path="/billing/*" element={<BillingMicrofrontend />} />
                              <Route path="/analytics/*" element={<AnalyticsMicrofrontend />} />
                              <Route path="/settings/*" element={<SettingsMicrofrontend />} />
                            </Routes>
                          </Suspense>
                        </Layout>
                      </DesignSystemProvider>
                    </OfflineProvider>
                  </LocalizationProvider>
                </WebSocketProvider>
              </RBACProvider>
            </TenantProvider>
          </AuthProvider>
        </BrowserRouter>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}

export default App;
