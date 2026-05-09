// clients/web-portal/microfrontends/shell/components/RBACGuard.tsx
import React, { useContext, createContext, useMemo } from 'react';
import { useAuth } from '../hooks/useAuth';
import { useTenant } from '../hooks/useTenant';

interface Permission {
  resource: string;
  action: string;
  scope?: string;
}

interface RBACContextType {
  permissions: Permission[];
  roles: string[];
  hasPermission: (resource: string, action: string) => boolean;
  hasRole: (role: string) => boolean;
  hasAnyRole: (roles: string[]) => boolean;
  hasAllRoles: (roles: string[]) => boolean;
}

const RBACContext = createContext<RBACContextType | undefined>(undefined);

export function RBACProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();
  const { tenant } = useTenant();

  const rbac = useMemo(() => {
    const permissions = user?.permissions || [];
    const roles = user?.roles || [];

    return {
      permissions,
      roles,
      hasPermission: (resource: string, action: string) => {
        return permissions.some(
          p => p.resource === resource && 
               (p.action === action || p.action === '*')
        );
      },
      hasRole: (role: string) => roles.includes(role),
      hasAnyRole: (requiredRoles: string[]) => 
        requiredRoles.some(r => roles.includes(r)),
      hasAllRoles: (requiredRoles: string[]) => 
        requiredRoles.every(r => roles.includes(r)),
    };
  }, [user]);

  return (
    <RBACContext.Provider value={rbac}>
      {children}
    </RBACContext.Provider>
  );
}

export function useRBAC() {
  const context = useContext(RBACContext);
  if (!context) {
    throw new Error('useRBAC must be used within RBACProvider');
  }
  return context;
}

interface PermissionGuardProps {
  resource: string;
  action: string;
  fallback?: React.ReactNode;
  children: React.ReactNode;
}

export function PermissionGuard({ 
  resource, 
  action, 
  fallback = null, 
  children 
}: PermissionGuardProps) {
  const { hasPermission } = useRBAC();

  if (!hasPermission(resource, action)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}

interface RoleGuardProps {
  roles: string[];
  requireAll?: boolean;
  fallback?: React.ReactNode;
  children: React.ReactNode;
}

export function RoleGuard({ 
  roles, 
  requireAll = false, 
  fallback = null, 
  children 
}: RoleGuardProps) {
  const { hasAnyRole, hasAllRoles } = useRBAC();

  const hasAccess = requireAll ? hasAllRoles(roles) : hasAnyRole(roles);

  if (!hasAccess) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}

interface TenantGuardProps {
  allowedTiers?: string[];
  fallback?: React.ReactNode;
  children: React.ReactNode;
}

export function TenantGuard({ 
  allowedTiers, 
  fallback = null, 
  children 
}: TenantGuardProps) {
  const { tenant } = useTenant();

  if (allowedTiers && !allowedTiers.includes(tenant.tier)) {
    return (
      <div className="p-4 bg-amber-50 border border-amber-200 rounded-lg">
        <p className="text-amber-800">
          This feature requires {allowedTiers.join(' or ')} tier.
        </p>
      </div>
    );
  }

  return <>{children}</>;
}
