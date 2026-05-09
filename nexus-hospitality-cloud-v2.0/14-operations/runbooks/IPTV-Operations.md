# NHC IPTV Operations Runbook
## Version 1.0 | SRE Team

---

## 1. Service Overview

| Service | Endpoint | Health Check | Critical |
|---------|----------|-------------|----------|
| Session Manager | :8081 | /health | Yes |
| Stream Controller | :8082 / :9090 | /health | Yes |
| EPG Service | :8084 | /health | Yes |
| DRM Service | :8085 | /health | Yes |
| Content Catalog | :8083 | /health | No |
| Casting Service | :8086 | /health | No |

## 2. Common Issues & Resolution

### 2.1 Stream Not Starting

**Symptoms**: Guest reports "Stream not loading" or black screen

**Diagnosis**:
```bash
# Check session validity
kubectl exec -it deploy/iptv-session-manager --   curl -H "X-Tenant-ID: $TENANT"   http://localhost:8081/api/v1/iptv/sessions/$SESSION_ID

# Check stream URL generation
kubectl logs deploy/iptv-stream-controller --tail=100 | grep $SESSION_ID

# Check CDN health
curl -I https://cf.nhc-cdn.com/health
```

**Resolution Steps**:
1. Verify guest entitlements in database
2. Check if session is active (not expired)
3. Verify CDN endpoint health
4. Regenerate stream URL if expired
5. Escalate to DRM team if license issue

### 2.2 High Stream Error Rate

**Symptoms**: Alert: `iptv_stream_errors_total > 10/min`

**Diagnosis**:
```bash
# Check error types
kubectl logs deploy/iptv-session-manager | grep "error" | jq -r '.error_type' | sort | uniq -c

# Check CDN latency
kubectl logs deploy/iptv-stream-controller | grep "latency"

# Check Redis connection
kubectl exec -it deploy/iptv-session-manager -- redis-cli ping
```

**Resolution**:
1. Identify error type (entitlement, DRM, CDN)
2. If CDN issue: Failover to backup CDN
3. If entitlement issue: Check PMS sync status
4. If DRM issue: Restart DRM service pods
5. Scale session manager if load-related

### 2.3 Casting Not Working

**Symptoms**: Guest cannot cast to room TV

**Diagnosis**:
```bash
# Check casting session
kubectl logs deploy/iptv-casting-service | grep $ROOM_ID

# Verify network isolation
kubectl exec -it deploy/iptv-casting-service --   nmap -p 8009 $TV_IP  # Chromecast port

# Check mDNS discovery
kubectl exec -it deploy/iptv-casting-service --   avahi-browse -a | grep Chromecast
```

**Resolution**:
1. Verify guest is on correct WiFi network
2. Check VLAN isolation (should be room-specific)
3. Restart casting service if stuck
4. Manual pairing via QR code as fallback
5. Escalate to network team if WiFi issue

### 2.4 Emergency Broadcast Failure

**Symptoms**: Emergency alert not displaying on TVs

**Immediate Actions**:
```bash
# Force broadcast via API
kubectl exec -it deploy/iptv-session-manager -- curl -X POST   -H "Content-Type: application/json"   -d '{"level":"critical","message":"EVACUATE","property_ids":["'$PROPERTY_ID'"]}'   http://localhost:8081/api/v1/iptv/emergency/broadcast

# Verify delivery
kubectl logs deploy/iptv-session-manager | grep "EMERGENCY"
```

**Escalation**: If automated broadcast fails, instruct front desk to:
1. Use manual override on property management console
2. Call rooms directly via PBX
3. Activate physical alarm systems

## 3. Scaling Procedures

### 3.1 Manual Scale-Up

```bash
# Scale session manager
kubectl scale deploy/iptv-session-manager --replicas=20 -n nexus-iptv

# Scale stream controller
kubectl scale deploy/iptv-stream-controller --replicas=15 -n nexus-iptv

# Verify
kubectl get pods -n nexus-iptv -l app=iptv-session-manager
```

### 3.2 Emergency Scale-Up (Major Event)

```bash
# Scale all IPTV services
kubectl patch hpa iptv-session-manager-hpa -n nexus-iptv   --patch '{"spec":{"maxReplicas":200}}'

# Enable additional CDN endpoints
kubectl patch configmap cdn-config -n nexus-iptv   --patch '{"data":{"endpoints":"cf,fastly,akamai,aws"}}'

# Pre-warm cache
kubectl exec -it deploy/cache-warmer -- /warm-cache.sh
```

## 4. Backup & Recovery

### 4.1 Database Backup

```bash
# CockroachDB backup
cockroach sql --url $DB_URL   --execute "BACKUP DATABASE iptv TO 's3://nhc-backups/iptv/$(date +%Y%m%d)'"

# Redis backup
kubectl exec -it redis-master-0 -- redis-cli BGSAVE
kubectl cp redis-master-0:/data/dump.rdb /backups/redis-$(date +%Y%m%d).rdb
```

### 4.2 Disaster Recovery

| Scenario | RTO | RPO | Procedure |
|----------|-----|-----|-----------|
| Single pod failure | < 1min | 0 | Auto-restart by K8s |
| Node failure | < 5min | 0 | Pod rescheduling |
| AZ failure | < 15min | 0 | Multi-AZ failover |
| Region failure | < 1hour | < 5min | DR region activation |
| Data corruption | < 4hours | < 1hour | Point-in-time restore |

## 5. Maintenance Windows

| Activity | Frequency | Duration | Impact |
|----------|-----------|----------|--------|
| Certificate rotation | Monthly | 5min | None (hot reload) |
| OS patching | Monthly | 30min | Rolling, no impact |
| K8s upgrade | Quarterly | 2hours | Rolling |
| Database maintenance | Quarterly | 4hours | Read-only window |
| Content catalog update | Weekly | 1hour | None |

## 6. Escalation Matrix

| Severity | Response Time | Escalation Path |
|----------|--------------|-----------------|
| P1 (Critical) | 5 min | On-call -> SRE Lead -> CTO |
| P2 (High) | 15 min | On-call -> Team Lead |
| P3 (Medium) | 1 hour | Ticket queue |
| P4 (Low) | 4 hours | Backlog |

## 7. Contact Information

| Role | Contact | Slack |
|------|---------|-------|
| On-call SRE | PagerDuty | #sre-oncall |
| IPTV Team Lead | John Smith | #iptv-team |
| Platform Architect | Jane Doe | #platform-arch |
| Security Team | security@nexushc.com | #security |
| CDN Provider | Cloudflare Enterprise | N/A |
