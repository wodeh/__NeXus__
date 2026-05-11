# API Integration Guide

The backend API hooks are ready in `src/hooks/useApi.ts`.

## Quick Integration

Replace mock data in any page with:

```tsx
import { useIPTVChannels, useSmartLocks, useRooms } from "@/hooks/useApi";

// In your component:
const { data: channels, loading, error, refetch } = useIPTVChannels();
const displayChannels = channels ?? FALLBACK_CHANNELS;
```

## Available Hooks

### IPTV
- `useIPTVChannels()` — GET /tenants/{id}/iptv/channels
- `useIPTVContent()` — GET /tenants/{id}/iptv/content
- `useIPTVRoomBindings()` — GET /tenants/{id}/iptv/room-bindings
- `useIPTVWelcomeScreen(roomId)` — GET /tenants/{id}/iptv/rooms/{roomId}/welcome
- `useIPTVAnalytics()` — GET /tenants/{id}/iptv/analytics
- `createIPTVChannel(channel)` — POST
- `updateIPTVChannel(id, patch)` — PATCH
- `deleteIPTVChannel(id)` — DELETE

### Smart Locks
- `useSmartLocks()` — GET /tenants/{id}/locks
- `useLockOverview()` — GET /tenants/{id}/locks/overview
- `useLockEvents(lockId)` — GET /tenants/{id}/locks/{lockId}/events
- `useAccessCodes(lockId)` — GET /tenants/{id}/locks/{lockId}/access-codes
- `useLockByRoom(roomId)` — GET /tenants/{id}/locks/by-room/{roomId}
- `remoteUnlock(lockId)` — POST
- `remoteLock(lockId)` — POST
- `createAccessCode(lockId, code)` — POST
- `revokeAccessCode(lockId, codeId)` — POST

### Command Center / Rooms
- `useRooms(propertyId?)` — GET /tenants/{id}/properties/{pid}/rooms
- `useDashboardStats()` — GET (aggregated stats endpoint)

### Reservations
- `useReservations()` — GET /tenants/{id}/reservations
- `useReservation(id)` — GET /tenants/{id}/reservations/{id}
- `checkInReservation(id)` — POST
- `checkOutReservation(id)` — POST
- `cancelReservation(id)` — PATCH

## Auth

The API client reads `token` and `tenantId` from localStorage. Set these on login:

```ts
localStorage.setItem("token", jwtToken);
localStorage.setItem("tenantId", tenantSlug);
```

## Environment

Set `NEXT_PUBLIC_API_URL` in `.env.local`:

```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Fallback Pattern

All pages currently use mock data. To wire to real APIs:

1. Import the hook
2. Use `data ?? fallback` for seamless degradation
3. Show loading spinner when `loading` is true
4. Show error toast when `error` is present
5. Add a "Refresh" button calling `refetch()`

## Example: Wiring IPTV Channels

```tsx
import { useIPTVChannels } from "@/hooks/useApi";

export default function IPTVPage() {
  const { data: apiChannels, loading, error, refetch } = useIPTVChannels();
  const channels = apiChannels ?? mockChannels;

  return (
    <div>
      {loading && <Spinner />}
      {error && <ErrorBanner message={error.message} onRetry={refetch} />}
      <ChannelGrid channels={channels} />
    </div>
  );
}
```
