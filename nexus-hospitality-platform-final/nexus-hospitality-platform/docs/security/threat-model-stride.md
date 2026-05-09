# NHC Threat Model & Security Architecture
## STRIDE Analysis + Attack Surface Mapping

---

## 1. STRIDE ANALYSIS

### 1.1 Spoofing (Identity)

| Component | Threat | Likelihood | Impact | Mitigation |
|-----------|--------|-----------|--------|------------|
| API Gateway | Fake JWT tokens | Medium | Critical | RS256 signing, key rotation, JWKS endpoint |
| IPTV Session | Session hijacking | Medium | High | Short TTL, device fingerprinting, IP binding |
| Smart Lock | Cloned key cards | Low | Critical | Cryptographic keys, rotation, audit logging |
| Mobile App | Device spoofing | Medium | Medium | Device attestation, SafetyNet/App Attest |
| Staff Portal | Credential stuffing | High | High | MFA, rate limiting, breach detection |

### 1.2 Tampering (Data Integrity)

| Component | Threat | Likelihood | Impact | Mitigation |
|-----------|--------|-----------|--------|------------|
| Stream Manifest | Modified to bypass DRM | Medium | High | Signed manifests, HMAC verification |
| Pricing Data | Rate manipulation | Low | Critical | Immutable audit log, approval workflows |
| Guest Profile | Data poisoning | Low | Medium | Input validation, schema enforcement |
| IoT Commands | Hijacked device control | Medium | Critical | mTLS, command signing, device attestation |

### 1.3 Repudiation (Non-repudiation)

| Component | Threat | Likelihood | Impact | Mitigation |
|-----------|--------|-----------|--------|------------|
| Financial Transactions | Denied bookings | Low | Critical | Immutable ledger, digital signatures |
| Staff Actions | Denied access grants | Low | High | Comprehensive audit logging, WORM storage |
| Stream Access | Denied content viewing | Low | Medium | Session logs, watermarking |

### 1.4 Information Disclosure

| Component | Threat | Likelihood | Impact | Mitigation |
|-----------|--------|-----------|--------|------------|
| Guest PII | Data breach | Medium | Critical | Encryption at rest, field-level encryption |
| Payment Data | PCI breach | Low | Critical | Tokenization, HSM, network segmentation |
| Stream Content | Content piracy | High | High | DRM, forensic watermarking, geo-blocking |
| Analytics Data | Competitive intelligence | Medium | Medium | Data anonymization, access controls |

### 1.5 Denial of Service

| Component | Threat | Likelihood | Impact | Mitigation |
|-----------|--------|-----------|--------|------------|
| Booking Engine | Bot attacks | High | High | WAF, CAPTCHA, rate limiting, bot detection |
| IPTV Streaming | Bandwidth exhaustion | Medium | High | CDN, traffic shaping, QoS |
| API Gateway | DDoS | High | Critical | AWS Shield, Cloudflare, rate limiting |
| IoT Gateway | Device flooding | Medium | Medium | Device quotas, anomaly detection |

### 1.6 Elevation of Privilege

| Component | Threat | Likelihood | Impact | Mitigation |
|-----------|--------|-----------|--------|------------|
| RBAC | Role escalation | Low | Critical | ABAC, least privilege, regular audits |
| Container | Container escape | Low | Critical | gVisor, seccomp, read-only rootfs |
| Kubernetes | Pod privilege escalation | Low | Critical | OPA Gatekeeper, Pod Security Standards |
| Database | SQL injection | Medium | Critical | Parameterized queries, WAF, input validation |

---

## 2. ZERO-TRUST ARCHITECTURE

### Core Principles
1. Never Trust, Always Verify
2. Least Privilege
3. Assume Breach
4. Verify Explicitly
5. Use Least Privilege Access

### Implementation Layers
- LAYER 1: IDENTITY (Keycloak, MFA, Biometric)
- LAYER 2: DEVICE (Registration, Attestation, MDM)
- LAYER 3: NETWORK (Micro-segmentation, mTLS, DDoS)
- LAYER 4: APPLICATION (OPA, WAF, Rate Limiting)
- LAYER 5: DATA (Encryption, Tokenization, DLP)

---

## 3. INCIDENT RESPONSE PLAYBOOKS

### Data Breach Response (72h GDPR)
- T+0-15min: Detection, automated containment, notification
- T+15-60min: Scope assessment, evidence preservation
- T+1-4h: Credential revocation, secret rotation, IP blocking
- T+4-24h: Vulnerability patching, clean verification
- T+24-72h: Service restoration, regulatory notification

### DDoS Response
- T+0: Automated detection (Cloudflare/AWS Shield)
- T+30s: Rate limiting activated
- T+1m: CDN absorption scaled
- T+2m: Blackhole routing if needed
- T+10m: Tenant communication

### Ransomware Response
- T+0: Detection (Falco/EDR)
- T+1m: Isolate affected systems
- T+15m: Assess encryption scope
- T+1h: Restore from immutable backups
- T+24h: Full service restoration
