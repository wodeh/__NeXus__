# Contributing to Nexus Hospitality Platform

## Development Workflow

1. **Fork & Branch**: Create feature branch from `main`
   ```bash
   git checkout -b feature/ABC-123-description
   ```

2. **Code Standards**:
   - Go: Follow Effective Go + golangci-lint rules
   - TypeScript: ESLint + Prettier + strict mode
   - SQL: Use migrations, never modify existing migrations
   - Terraform: `terraform fmt` + `tflint`

3. **Testing**:
   - Unit tests: >80% coverage
   - Integration tests: Required for API changes
   - E2E tests: Required for UI changes
   - Chaos tests: Required for infrastructure changes

4. **Security**:
   - All commits signed (GPG)
   - Snyk scan passed
   - Trivy scan passed
   - No secrets in code (use external-secrets)

5. **Pull Request**:
   - Fill PR template completely
   - Link to Jira ticket
   - Request review from CODEOWNERS
   - CI must pass (see `.github/workflows/`)

6. **Deployment**:
   - Canary deployment: 5% → 25% → 50% → 100%
   - Automated rollback on error rate >1%
   - Post-deployment monitoring: 30 minutes

## Code Review Checklist

- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Security implications considered
- [ ] Performance impact assessed
- [ ] Backward compatibility maintained
- [ ] Observability hooks added (metrics, logs, traces)
- [ ] Feature flags implemented for risky changes

## Commit Message Format

```
type(scope): subject

body (optional)

footer (optional)
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`

Example:
```
feat(iptv): add SCTE-35 ad insertion support

Implements splice_insert and time_signal marker types.
Supports pre-roll and mid-roll ad breaks.

Closes: ABC-123
```

## Emergency Hotfix Process

1. Branch from `main`: `hotfix/ABC-456-critical-fix`
2. Fix with minimal changes
3. Fast-track review (1 approval)
4. Deploy directly to production with manual approval
5. Merge back to `main` and `release/*` branches

## Support

- Slack: `#nexus-dev`
- Wiki: https://wiki.nexus-platform.com
- Office Hours: Tuesdays 10:00 UTC
