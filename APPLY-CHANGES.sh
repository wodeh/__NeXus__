#!/bin/bash
# NeXus Phase 1 Changes — Apply Script
# Run this from your __NeXus__ repo root

set -e

echo "=== NeXus Phase 1 Security & Stability Patches ==="
echo ""

# Check we're in the right place
if [ ! -f "README.md" ] || [ ! -d "services" ]; then
    echo "ERROR: Run this script from your __NeXus__ repo root directory"
    echo "Expected: README.md and services/ directory"
    exit 1
fi

echo "[1/4] Backing up original files..."
mkdir -p .nexus-backup-phase1
cp services/backend-core/internal/server/auth_handlers.go .nexus-backup-phase1/ 2>/dev/null || true
cp services/backend-core/internal/server/server.go .nexus-backup-phase1/ 2>/dev/null || true
cp services/backend-core/internal/server/cors.go .nexus-backup-phase1/ 2>/dev/null || true
cp services/backend-core/internal/auth/jwt.go .nexus-backup-phase1/ 2>/dev/null || true
cp services/backend-core/internal/auth/tenant.go .nexus-backup-phase1/ 2>/dev/null || true
cp services/backend-core/internal/repository/repository.go .nexus-backup-phase1/ 2>/dev/null || true
cp services/backend-core/internal/repository/reservation.go .nexus-backup-phase1/ 2>/dev/null || true
cp services/backend-core/internal/repository/housekeeping.go .nexus-backup-phase1/ 2>/dev/null || true
cp "services/backend-core/sql/migrations/015_fix_rls_policies.up.sql" .nexus-backup-phase1/ 2>/dev/null || true
cp apps/web/src/lib/auth.tsx .nexus-backup-phase1/ 2>/dev/null || true
cp apps/web/src/lib/api.ts .nexus-backup-phase1/ 2>/dev/null || true
cp apps/web/src/hooks/useApi.ts .nexus-backup-phase1/ 2>/dev/null || true
cp "apps/web/src/app/(dashboard)/layout.tsx" .nexus-backup-phase1/ 2>/dev/null || true
cp apps/web/src/components/Topbar.tsx .nexus-backup-phase1/ 2>/dev/null || true
cp "apps/web/src/app/(dashboard)/settings/page.tsx" .nexus-backup-phase1/ 2>/dev/null || true
cp .github/workflows/ci-pms.yml .nexus-backup-phase1/ 2>/dev/null || true
cp .github/workflows/release-orchestrator.yml .nexus-backup-phase1/ 2>/dev/null || true
echo "  Backup saved to .nexus-backup-phase1/"

echo ""
echo "[2/4] Copying new files..."
cp -r nexus-patch/services/backend-core/internal/server/ratelimit.go services/backend-core/internal/server/ 2>/dev/null
cp -r nexus-patch/services/backend-core/internal/server/validator.go services/backend-core/internal/server/ 2>/dev/null
cp -r nexus-patch/services/backend-core/internal/repository/helpers.go services/backend-core/internal/repository/ 2>/dev/null
cp -r nexus-patch/services/backend-core/internal/auth/tenant_test.go services/backend-core/internal/auth/ 2>/dev/null
cp -r nexus-patch/apps/web/src/components/ErrorBoundary.tsx apps/web/src/components/ 2>/dev/null
cp -r nexus-patch/.github/workflows/ci-backend-core.yml .github/workflows/ 2>/dev/null
cp -r nexus-patch/.github/workflows/ci-ai-platform.yml .github/workflows/ 2>/dev/null
cp -r nexus-patch/.github/workflows/ci-iot-gateway.yml .github/workflows/ 2>/dev/null
cp -r nexus-patch/.github/workflows/ci-iptv-middleware.yml .github/workflows/ 2>/dev/null
echo "  8 new files copied"

echo ""
echo "[3/4] Copying modified files..."
cp nexus-patch/services/backend-core/internal/server/auth_handlers.go services/backend-core/internal/server/
cp nexus-patch/services/backend-core/internal/server/server.go services/backend-core/internal/server/
cp nexus-patch/services/backend-core/internal/server/cors.go services/backend-core/internal/server/
cp nexus-patch/services/backend-core/internal/auth/jwt.go services/backend-core/internal/auth/
cp nexus-patch/services/backend-core/internal/auth/tenant.go services/backend-core/internal/auth/
cp nexus-patch/services/backend-core/internal/repository/repository.go services/backend-core/internal/repository/
cp nexus-patch/services/backend-core/internal/repository/reservation.go services/backend-core/internal/repository/
cp nexus-patch/services/backend-core/internal/repository/housekeeping.go services/backend-core/internal/repository/
cp "nexus-patch/services/backend-core/sql/migrations/015_fix_rls_policies.up.sql" "services/backend-core/sql/migrations/"
cp nexus-patch/apps/web/src/lib/auth.tsx apps/web/src/lib/
cp nexus-patch/apps/web/src/lib/api.ts apps/web/src/lib/
cp nexus-patch/apps/web/src/hooks/useApi.ts apps/web/src/hooks/
cp "nexus-patch/apps/web/src/app/(dashboard)/layout.tsx" "apps/web/src/app/(dashboard)/"
cp nexus-patch/apps/web/src/components/Topbar.tsx apps/web/src/components/
cp "nexus-patch/apps/web/src/app/(dashboard)/settings/page.tsx" "apps/web/src/app/(dashboard)/settings/"
cp nexus-patch/.github/workflows/ci-pms.yml .github/workflows/
cp nexus-patch/.github/workflows/release-orchestrator.yml .github/workflows/
echo "  17 modified files copied"

echo ""
echo "[4/4] Cleaning up..."
rm -rf nexus-hospitality-platform-final/ 2>/dev/null || true
echo "  Removed duplicate nexus-hospitality-platform-final/ directory"

echo ""
echo "=== DONE ==="
echo ""
echo "Next steps:"
echo "  1. Review changes: git diff --stat"
echo "  2. Run go mod tidy in each service:"
echo "     cd services/backend-core && go mod tidy"
echo "     cd ../pms-integration && go mod tidy"
echo "     cd ../ai-platform && go mod tidy"
echo "     cd ../iot-gateway && go mod tidy"
echo "     cd ../iptv-middleware && go mod tidy"
echo "  3. Test: make dev"
echo "  4. Stage, branch, commit:"
echo "     git checkout -b phase1/security-stability"
echo "     git add -A"
echo "     git commit -m 'Phase 1: Security & Stability fixes'"
echo "  5. Push branch (NOT main):"
echo "     git push -u origin phase1/security-stability"
echo ""
echo "  6. Open PR on GitHub:"
echo "     https://github.com/wodeh/__NeXus__/pull/new/phase1/security-stability"
echo "     Never push directly to main — always review through a PR."
echo ""
echo "Backup location: .nexus-backup-phase1/"
echo "To restore: cp .nexus-backup-phase1/* <original locations>"
