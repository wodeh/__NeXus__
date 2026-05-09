# NHC Identity & Access Management Architecture

## 1. Identity Providers
- Keycloak (Primary) - OIDC, SAML, LDAP
- Azure AD (Enterprise) - SAML, SCIM, MFA
- Google Workspace (SMB) - OAuth, OIDC, SSO

## 2. RBAC Model

### Staff Roles
| Role | Permissions | Scope |
|------|------------|-------|
| System Admin | Full access | Global |
| Regional Manager | All permissions | Assigned regions |
| General Manager | All permissions | Single property |
| Front Desk Agent | Guest CRUD, check-in/out | Single property |
| Housekeeping Supervisor | Room status, maintenance | Single property |
| Revenue Manager | Rates, forecasts | Assigned properties |
| Night Auditor | End-of-day, reports | Single property |
| Concierge | Guest services | Single property |
| Read-Only | View only | Assigned properties |

### Guest Roles
| Role | Permissions |
|------|------------|
| Anonymous | View availability, rates |
| Registered | Book, manage reservations |
| Checked-in | Room controls, IPTV, services |
| Loyalty Member | Points, upgrades, offers |
| VIP | All services, priority support |

## 3. ABAC Policies (OPA)
- Time-based: Night auditor access 22:00-06:00 only
- Location-based: Front desk from property IP only
- Device-based: Admin requires trusted device + MFA
- Data sensitivity: Credit card masked without MFA
- Workflow: Rate changes >20% require approval

## 4. Service-to-Service (mTLS)
- Istio Citadel: Automatic cert rotation (24h TTL)
- SPIFFE identities: spiffe://nhc/ns/{namespace}/sa/{service}
- SVID validation on every request
- Rate limits per service identity

## 5. Vault Secrets Management
- HA cluster (5 nodes), Raft storage
- Auto-unseal: AWS KMS
- Dynamic secrets: DB creds (1h TTL), AWS STS (15m TTL)
- Encryption: AES-256-GCM, HSM-backed master key
- Hierarchy: global/tenants/{id}/services/{name}/infra/
