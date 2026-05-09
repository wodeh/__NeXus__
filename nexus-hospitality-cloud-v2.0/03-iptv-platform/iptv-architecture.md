# NEXUS HOSPITALITY IPTV PLATFORM (NHIPTV)
## Enterprise IPTV Middleware & Streaming Architecture v2.0

---

## 1. IPTV SYSTEM ARCHITECTURE

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        NHIPTV PLATFORM OVERVIEW                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                    CONTENT ACQUISITION LAYER                          │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │    │
│  │  │ Live TV  │ │ VOD      │ │ Catch-up │ │ Hotel    │ │ Premium  │   │    │
│  │  │ Sources  │ │ Catalog  │ │ TV       │ │ Content  │ │ Content  │   │    │
│  │  │ (DVB/IP) │ │ (Studios)│ │ (DVR)    │ │ (CMS)    │ │ (HBO,..) │   │    │
│  │  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘   │    │
│  │       └─────────────┴─────────────┴─────────────┴─────────────┘       │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                    │                                         │
│  ┌─────────────────────────────────▼─────────────────────────────────────┐   │
│  │                    MEDIA PROCESSING & TRANSCODING                       │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │  Ingestion  │  │ Transcoding │  │   DRM       │  │ Packaging   │  │   │
│  │  │  Servers    │──▶│  Cluster    │──▶│  Wrapping   │──▶│  (HLS/DASH) │  │   │
│  │  │  (FFmpeg)   │  │  (GPU)      │  │  (Widevine) │  │             │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                   │   │
│  │  │  Quality    │  │  Thumbnail  │  │  Subtitle   │                   │   │
│  │  │  Analysis   │  │  Generation │  │  Extraction │                   │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘                   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                         │
│  ┌─────────────────────────────────▼─────────────────────────────────────┐   │
│  │                    ORIGIN & CDN LAYER                                   │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │  Origin     │  │  CDN Edge   │  │  Regional   │  │  Property   │  │   │
│  │  │  Servers    │──▶│  Nodes      │──▶│  Cache      │──▶│  Edge       │  │   │
│  │  │  (Nginx)    │  │  (Cloudflare│  │  (Redis)    │  │  (Local)    │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                         │
│  ┌─────────────────────────────────▼─────────────────────────────────────┐   │
│  │                    IPTV MIDDLEWARE (K8s)                                │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │  Session    │  │  EPG        │  │  Content    │  │  Casting    │  │   │
│  │  │  Manager    │  │  Service    │  │  Discovery  │  │  Service    │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │  Guest      │  │  Billing    │  │  Analytics  │  │  AI Rec     │  │   │
│  │  │  Profile    │  │  Integration│  │  Engine     │  │  Engine     │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                         │
│  ┌─────────────────────────────────▼─────────────────────────────────────┐   │
│  │                    CLIENT LAYER                                         │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐    │   │
│  │  │ Android  │ │  webOS   │ │  Tizen   │ │  STB     │ │  Mobile  │    │   │
│  │  │  TV App  │ │  LG App  │ │Samsung App│ │  Linux   │ │  Casting │    │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘    │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐                  │   │
│  │  │  Tablet  │ │  In-Room │ │  Digital │ │  Kiosk   │                  │   │
│  │  │  App     │ │  Browser │ │  Signage │ │  Display │                  │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. IPTV MIDDLEWARE MICROSERVICES

### 2.1 Service Topology

```
┌─────────────────────────────────────────────────────────────────┐
│                    NHIPTV SERVICE MESH                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  Session    │  │  Stream     │  │  Content    │             │
│  │  Manager    │  │  Controller │  │  Catalog    │             │
│  │  (Go)       │  │  (Go)       │  │  (Go)       │             │
│  │  Port: 8081 │  │  Port: 8082 │  │  Port: 8083 │             │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘             │
│         │                │                │                    │
│  ┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐             │
│  │  EPG        │  │  DRM        │  │  Casting    │             │
│  │  Service    │  │  Service    │  │  Service    │             │
│  │  (Go)       │  │  (Go)       │  │  (Node.js)  │             │
│  │  Port: 8084 │  │  Port: 8085 │  │  Port: 8086 │             │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘             │
│         │                │                │                    │
│  ┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐             │
│  │  Guest      │  │  Billing    │  │  Analytics  │             │
│  │  Profile    │  │  Service    │  │  Service    │             │
│  │  (Go)       │  │  (Go)       │  │  (Python)   │             │
│  │  Port: 8087 │  │  Port: 8088 │  │  Port: 8089 │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  AI Rec     │  │  Voice      │  │  Emergency  │             │
│  │  Engine     │  │  Control    │  │  Broadcast  │             │
│  │  (Python)   │  │  (Python)   │  │  (Go)       │             │
│  │  Port: 8090 │  │  Port: 8091 │  │  Port: 8092 │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Session Manager Service

Responsible for:
- Guest session lifecycle
- Device authentication
- Concurrent stream limits
- Geo-fencing
- Parental controls
- Viewing history

### 2.3 Stream Controller Service

Responsible for:
- Stream URL generation
- CDN routing optimization
- Quality adaptation
- Bandwidth management
- Failover handling
- QoS monitoring

### 2.4 Content Catalog Service

Responsible for:
- VOD metadata management
- Content categorization
- Search and discovery
- Content licensing
- Availability windows
- Regional restrictions

### 2.5 EPG Service

Responsible for:
- Electronic program guide
- Schedule management
- Channel mapping
- Program metadata
- Recording scheduling
- Reminder management

### 2.6 DRM Service

Responsible for:
- License issuance
- Content encryption
- Device authorization
- License renewal
- Revocation
- Watermarking

### 2.7 Casting Service

Responsible for:
- Chromecast discovery
- AirPlay pairing
- QR code generation
- Session bridging
- Quality negotiation
- Security isolation

### 2.8 AI Recommendation Engine

Responsible for:
- Content recommendations
- Personalized EPG
- Trending content
- Similar content
- Guest preference learning
- A/B testing

---

## 3. STREAMING PROTOCOLS & FORMATS

### 3.1 Protocol Matrix

| Content Type | Primary Protocol | Fallback | Latency | Use Case |
|-------------|-----------------|----------|---------|----------|
| Live TV | HLS (CMAF) | DASH | 6-10s | Broadcast |
| Live TV (Low Latency) | LL-HLS | LL-DASH | 2-4s | Sports |
| VOD | HLS | DASH | N/A | Movies, Shows |
| Catch-up | HLS | DASH | N/A | Time-shifted |
| Hotel Content | HLS | Progressive | N/A | Promos, Info |
| Casting | WebRTC | HLS | <1s | Screen mirroring |
| Emergency | Multicast UDP | Unicast | <500ms | Alerts |

### 3.2 Transcoding Profiles

| Profile | Resolution | Bitrate | Codec | Audio |
|---------|-----------|---------|-------|-------|
| 4K UHD | 3840x2160 | 15 Mbps | H.265/AV1 | AAC 5.1 |
| 1080p HD | 1920x1080 | 8 Mbps | H.264 | AAC 2.0 |
| 720p HD | 1280x720 | 4 Mbps | H.264 | AAC 2.0 |
| 480p SD | 854x480 | 2 Mbps | H.264 | AAC 2.0 |
| 360p SD | 640x360 | 1 Mbps | H.264 | AAC 2.0 |
| 240p LD | 426x240 | 500 Kbps | H.264 | AAC 2.0 |
| Audio Only | — | 128 Kbps | — | AAC 2.0 |

---

## 4. DRM ARCHITECTURE

### 4.1 Multi-DRM Strategy

| DRM System | Browser | Android | iOS | Smart TV | Use Case |
|-----------|---------|---------|-----|----------|----------|
| **Widevine** | Chrome, FF | Native | — | Android TV, Tizen | Primary |
| **FairPlay** | Safari | — | Native | tvOS, webOS | Apple ecosystem |
| **PlayReady** | Edge, IE | — | — | Xbox, Samsung | Microsoft ecosystem |
| **AES-128** | All | All | All | All | Fallback |
| **ClearKey** | All | All | All | All | Non-DRM content |

### 4.2 License Flow

```
1. Client requests content playback
2. Stream Controller returns encrypted manifest
3. Client extracts initialization data (PSSH)
4. Client sends license request to DRM Service
5. DRM Service validates:
   - Device certificate
   - Content entitlement
   - Concurrent stream limit
   - Geo-location
   - Guest subscription tier
6. DRM Service generates license from Key Server
7. License delivered to client
8. Client decrypts and plays content
```

### 4.3 Forensic Watermarking

```
Per-guest watermark embedded in:
- A/B frame selection patterns
- LSB of select macroblocks
- Audio phase modulation
- Invisible to viewer
- Detectable by analysis
- Enables piracy tracing to specific guest/room
```

---

## 5. GUEST CASTING ARCHITECTURE

### 5.1 Chromecast Integration

```
Guest Room TV (with Chromecast built-in or dongle)
    │
    ├─ Discovery: mDNS/Bonjour broadcast
    │
    ├─ Pairing: QR code scan or manual PIN
    │
    ├─ Authentication: JWT token via secure channel
    │
    ├─ Session: Isolated VLAN per room
    │
    └─ Streaming: Guest device -> Chromecast (local WiFi)
                   NOT through internet (bandwidth efficient)
```

### 5.2 AirPlay Integration

```
Guest Room TV (with AirPlay 2 support)
    │
    ├─ Discovery: AirPlay beacon
    │
    ├─ Pairing: QR code or Apple HomeKit
    │
    ├─ Authentication: Apple ID verification + hotel token
    │
    ├─ Session: Encrypted AirPlay stream
    │
    └─ Streaming: Direct peer-to-peer or via AP
```

### 5.3 Security Isolation

```
Each room gets isolated network segment:
- VLAN per room (or per floor)
- mDNS reflection blocked between rooms
- Casting limited to same-room devices
- Auto-disconnect on checkout
- Bandwidth limit per casting session
- Content filtering (block adult content in family rooms)
```

---

## 6. SMART TV PLATFORM SUPPORT

### 6.1 Platform Matrix

| Platform | OS Version | App Framework | DRM | Casting | Status |
|----------|-----------|---------------|-----|---------|--------|
| **Android TV** | 10+ | Android SDK | Widevine | Chromecast | ✅ Native |
| **Google TV** | 12+ | Android SDK | Widevine | Chromecast | ✅ Native |
| **LG webOS** | 4.0+ | webOS SDK | Widevine | AirPlay | ✅ Native |
| **Samsung Tizen** | 5.0+ | Tizen SDK | PlayReady | AirPlay | ✅ Native |
| **Roku TV** | 10.0+ | BrightScript | Widevine | Roku Cast | ⚠️ Planned |
| **Amazon Fire TV** | 7+ | Android SDK | Widevine | Fling | ⚠️ Planned |
| **Apple tvOS** | 14+ | SwiftUI | FairPlay | AirPlay | ✅ Native |
| **Philips Android** | 9+ | Android SDK | Widevine | Chromecast | ✅ Native |
| **Sony Android** | 9+ | Android SDK | Widevine | Chromecast | ✅ Native |

### 6.2 Hospitality TV Specifics

| Feature | Samsung LYNK | LG Pro:Centric | Philips CMND | NHC Unified |
|---------|-------------|----------------|--------------|-------------|
| Remote management | ✅ | ✅ | ✅ | ✅ |
| Welcome screen | ✅ | ✅ | ✅ | ✅ |
| PMS integration | ⚠️ Limited | ⚠️ Limited | ⚠️ Limited | ✅ Deep |
| Billing integration | ❌ | ❌ | ❌ | ✅ Native |
| AI personalization | ❌ | ❌ | ❌ | ✅ Native |
| Casting | ⚠️ Basic | ⚠️ Basic | ⚠️ Basic | ✅ Advanced |
| Multi-property | ❌ | ❌ | ❌ | ✅ Native |
| API access | ❌ | ❌ | ❌ | ✅ Full |

---

## 7. EMERGENCY BROADCAST SYSTEM

### 7.1 Architecture

```
Emergency Alert Sources:
- National alert systems (IPAWS, EAS)
- Hotel security systems
- Fire alarm systems
- Weather alerts
- Manual override (front desk)

Distribution:
1. Emergency Service receives alert
2. Validates and prioritizes
3. Overrides ALL TV screens
4. Forces channel change to emergency channel
5. Displays alert with audio
6. Logs delivery confirmation
7. Auto-resumes after all-clear
```

### 7.2 Priority Levels

| Level | Type | Behavior | Override |
|-------|------|----------|----------|
| 1 | Critical (Fire, Active Shooter) | Full screen, audio, all devices | Cannot dismiss |
| 2 | Urgent (Evacuation, Severe Weather) | Full screen, audio | Dismiss after 30s |
| 3 | Important (Lockdown, Utility) | Banner + audio | Dismiss after 10s |
| 4 | Advisory (Maintenance, Event) | Banner only | Dismiss immediately |

---

## 8. IPTV BILLING INTEGRATION

### 8.1 Charge Types

| Charge | Trigger | Amount | Posting |
|--------|---------|--------|---------|
| Premium VOD | Playback start | $9.99-$19.99 | Real-time to folio |
| PPV Event | Purchase | $29.99-$59.99 | Real-time to folio |
| Adult Content | Playback start | $14.99 | Real-time to folio |
| Casting (premium) | Session start | $4.99/day | Real-time to folio |
| International channels | Channel tune | $2.99/day | Daily batch |
| Gaming | Session start | $5.99/hour | Real-time to folio |

### 8.2 Revenue Share Model

```
Content Provider: 60-70%
Platform (NHC):  15-20%
Hotel:           10-15%
Payment Processor: 2-3%

Example: $9.99 VOD rental
- Studio: $6.50
- NHC: $1.75
- Hotel: $1.25
- Stripe: $0.30 + 2.9%
```

---

## 9. AI-POWERED IPTV FEATURES

### 9.1 Personalized Content Discovery

```
Input Signals:
- Guest demographics (from PMS)
- Viewing history (anonymized)
- Time of day
- Day of week
- Weather
- Local events
- Language preference
- Content ratings preference

AI Model: Collaborative filtering + Content-based + Contextual
Output: Personalized EPG, Recommended VOD, Trending for you
```

### 9.2 Voice-Controlled TV

```
Wake Word: "Hey Nexus" or "Hey Hotel"

Commands:
- "Turn on CNN"
- "Find action movies"
- "Order room service"
- "What's the weather?"
- "Set alarm for 7 AM"
- "Checkout"

Architecture:
Guest Voice -> TV Mic -> Edge AI (Whisper) -> Intent Classification -> Action
```

### 9.3 Dynamic Welcome Screens

```
Content Personalization:
- Guest name and loyalty status
- Weather and local time
- Upcoming reservations (spa, restaurant)
- Personalized offers
- Local recommendations
- Hotel amenities highlight
- Language-matched content

Rendering: HTML5 overlay on TV boot
Update frequency: Real-time via WebSocket
```

---

## 10. IPTV OBSERVABILITY

### 10.1 QoS Metrics

| Metric | Target | Alert Threshold |
|--------|--------|----------------|
| Stream start time | < 2s | > 3s |
| Rebuffer ratio | < 0.5% | > 1% |
| Bitrate adaptation | < 5s | > 10s |
| CDN cache hit ratio | > 95% | < 90% |
| DRM license latency | < 500ms | > 1s |
| EPG load time | < 1s | > 2s |
| Casting setup time | < 5s | > 10s |
| Concurrent streams | Per plan | 90% of limit |

### 10.2 Monitoring Dashboards

- Real-time viewer count per channel
- Bandwidth utilization per property
- Content popularity heatmap
- Device type breakdown
- Error rate by platform
- Revenue per stream
- Guest satisfaction score

---

*Document Classification: Confidential — Technical Architecture*
*Version: 2.0 | Last Updated: 2026-05-07*
