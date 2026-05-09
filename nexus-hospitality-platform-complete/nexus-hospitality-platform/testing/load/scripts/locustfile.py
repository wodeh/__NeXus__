# testing/load/scripts/locustfile.py
# ============================================================
# LOCUST LOAD TESTS — Hospitality platform load simulation
# Simulates: guest check-ins, IPTV streaming, IoT telemetry,
# PMS operations, AI inference requests
# ============================================================

from locust import HttpUser, task, between, events
from locust.runners import MasterRunner
import random
import json
import time
from datetime import datetime, timedelta

class GuestUser(HttpUser):
    """Simulates a hotel guest using the guest portal"""
    wait_time = between(5, 30)
    weight = 70

    def on_start(self):
        # Guest authentication
        self.guest_id = f"guest_{self.user_id}"
        self.room_id = f"room_{random.randint(100, 9999)}"
        self.tenant_id = random.choice(["hotel-alpha", "hotel-beta", "resort-gamma"])

        response = self.client.post("/api/v1/auth/guest", json={
            "room_number": self.room_id,
            "last_name": "Smith",
            "check_in_date": (datetime.now() - timedelta(days=random.randint(0, 3))).isoformat(),
        })

        if response.status_code == 200:
            self.token = response.json()["token"]
            self.client.headers.update({"Authorization": f"Bearer {self.token}"})

    @task(10)
    def browse_guest_portal(self):
        """Browse hotel services and information"""
        endpoints = [
            "/api/v1/guest/services",
            "/api/v1/guest/dining/menu",
            "/api/v1/guest/spa/services",
            "/api/v1/guest/activities",
            "/api/v1/guest/local-guide",
        ]
        self.client.get(random.choice(endpoints))

    @task(5)
    def interact_with_iptv(self):
        """Simulate IPTV channel browsing and VOD"""
        # Channel zap
        channel_id = random.randint(1, 500)
        self.client.post("/api/v1/iptv/channel/change", json={
            "channel_id": channel_id,
            "room_id": self.room_id,
        })

        # Simulate viewing duration
        time.sleep(random.uniform(30, 300))

        # VOD request
        if random.random() < 0.3:
            self.client.post("/api/v1/iptv/vod/play", json={
                "content_id": f"movie_{random.randint(1, 10000)}",
                "room_id": self.room_id,
            })

    @task(3)
    def use_concierge_ai(self):
        """Interact with AI concierge"""
        prompts = [
            "What time is checkout?",
            "Recommend a restaurant nearby",
            "How do I connect to WiFi?",
            "Book a taxi to the airport",
            "What are the spa hours?",
            "Order room service",
            "Turn up the thermostat",
            "Set a wake-up call",
        ]

        response = self.client.post("/api/v1/ai/concierge", json={
            "prompt": random.choice(prompts),
            "guest_id": self.guest_id,
            "room_id": self.room_id,
            "language": random.choice(["en", "es", "fr", "de", "zh"]),
        })

        if response.status_code == 200:
            # Simulate reading response
            time.sleep(random.uniform(5, 15))

    @task(2)
    def control_smart_room(self):
        """Control IoT devices in room"""
        devices = ["thermostat", "lights", "curtains", "tv", "dnd"]
        device = random.choice(devices)

        actions = {
            "thermostat": {"temperature": random.randint(18, 26)},
            "lights": {"brightness": random.randint(0, 100), "scene": random.choice(["relax", "work", "sleep"])},
            "curtains": {"position": random.choice(["open", "closed", "half"])},
            "tv": {"power": random.choice(["on", "off"])},
            "dnd": {"enabled": random.choice([True, False])},
        }

        self.client.post(f"/api/v1/iot/room/{self.room_id}/device/{device}", json=actions[device])

    @task(1)
    def make_reservation_modification(self):
        """Modify existing reservation"""
        modifications = [
            {"late_checkout": True},
            {"room_upgrade_request": True},
            {"extend_stay": random.randint(1, 3)},
        ]

        self.client.patch("/api/v1/reservations/current", json=random.choice(modifications))

class FrontDeskStaff(HttpUser):
    """Simulates front desk staff operations"""
    wait_time = between(2, 10)
    weight = 20

    def on_start(self):
        self.tenant_id = random.choice(["hotel-alpha", "hotel-beta"])

        response = self.client.post("/api/v1/auth/staff", json={
            "employee_id": f"emp_{self.user_id}",
            "pin": "1234",
        })

        if response.status_code == 200:
            self.token = response.json()["token"]
            self.client.headers.update({"Authorization": f"Bearer {self.token}"})

    @task(15)
    def check_guest_in(self):
        """Simulate guest check-in process"""
        reservation_id = f"res_{random.randint(100000, 999999)}"

        # Validate reservation
        self.client.get(f"/api/v1/reservations/{reservation_id}")

        # Assign room
        room_response = self.client.post("/api/v1/rooms/assign", json={
            "reservation_id": reservation_id,
            "preferences": ["high_floor", "non_smoking"],
        })

        if room_response.status_code == 200:
            room_id = room_response.json()["room_id"]

            # Activate key card
            self.client.post("/api/v1/access/keycard/activate", json={
                "room_id": room_id,
                "guest_id": f"guest_{random.randint(1000, 9999)}",
                "checkout_date": (datetime.now() + timedelta(days=random.randint(1, 7))).isoformat(),
            })

            # Initialize IoT room state
            self.client.post(f"/api/v1/iot/room/{room_id}/initialize", json={
                "welcome_mode": True,
                "temperature": 22,
                "lighting_scene": "welcome",
            })

    @task(10)
    def process_housekeeping_update(self):
        """Update housekeeping status"""
        room_id = f"room_{random.randint(100, 9999)}"
        status = random.choice(["clean", "dirty", "inspected", "out_of_order"])

        self.client.patch(f"/api/v1/housekeeping/rooms/{room_id}", json={
            "status": status,
            "completed_by": f"hk_{random.randint(1, 50)}",
            "notes": random.choice(["", "Extra towels needed", "Light bulb replaced"]),
        })

    @task(8)
    def handle_maintenance_request(self):
        """Create or update maintenance requests"""
        if random.random() < 0.5:
            # Create new request
            self.client.post("/api/v1/maintenance/requests", json={
                "room_id": f"room_{random.randint(100, 9999)}",
                "category": random.choice(["plumbing", "electrical", "hvac", "appliance"]),
                "priority": random.choice(["low", "medium", "high", "emergency"]),
                "description": "Guest reported issue",
            })
        else:
            # Update existing
            self.client.patch(f"/api/v1/maintenance/requests/{random.randint(1, 1000)}", json={
                "status": random.choice(["assigned", "in_progress", "completed"]),
                "technician_id": f"tech_{random.randint(1, 20)}",
            })

    @task(5)
    def process_payment(self):
        """Process folio charges and payments"""
        folio_id = f"folio_{random.randint(100000, 999999)}"

        # Add charge
        self.client.post(f"/api/v1/folios/{folio_id}/charges", json={
            "category": random.choice(["room", "f&b", "spa", "minibar", "laundry"]),
            "amount": round(random.uniform(10, 500), 2),
            "description": "Guest service charge",
        })

        # Process payment
        if random.random() < 0.3:
            self.client.post(f"/api/v1/folios/{folio_id}/payments", json={
                "method": random.choice(["credit_card", "debit_card", "cash", "loyalty_points"]),
                "amount": round(random.uniform(50, 1000), 2),
            })

class IPTVViewer(HttpUser):
    """Simulates high-volume IPTV streaming"""
    wait_time = between(1, 5)
    weight = 50

    def on_start(self):
        self.room_id = f"room_{random.randint(100, 9999)}"
        self.tenant_id = random.choice(["hotel-alpha", "hotel-beta", "resort-gamma"])

        # Authenticate streaming session
        response = self.client.post("/api/v1/iptv/auth", json={
            "room_id": self.room_id,
            "device_id": f"tv_{random.randint(1000, 9999)}",
        })

        if response.status_code == 200:
            self.stream_token = response.json()["token"]

    @task(20)
    def stream_live_channel(self):
        """Simulate live TV streaming"""
        channel_id = random.randint(1, 500)
        profile = random.choice(["1080p", "720p", "480p", "360p"])

        # Request manifest
        manifest_response = self.client.get(
            f"/hls/channel_{channel_id}/{profile}/master.m3u8",
            headers={"X-Stream-Token": self.stream_token},
        )

        if manifest_response.status_code == 200:
            # Simulate segment requests
            for _ in range(random.randint(10, 60)):
                segment_id = random.randint(1, 10000)
                self.client.get(
                    f"/hls/channel_{channel_id}/{profile}/segment_{segment_id}.ts",
                    headers={"X-Stream-Token": self.stream_token},
                )
                time.sleep(6)  # 6-second segments

    @task(5)
    def stream_vod_content(self):
        """Simulate VOD playback"""
        content_id = f"movie_{random.randint(1, 10000)}"

        # Get DRM license
        self.client.post("/api/v1/iptv/drm/license", json={
            "content_id": content_id,
            "drm_type": random.choice(["widevine", "playready", "fairplay"]),
        })

        # Stream content
        self.client.get(
            f"/dash/{content_id}/manifest.mpd",
            headers={"X-Stream-Token": self.stream_token},
        )

    @task(3)
    def use_catchup_tv(self):
        """Access time-shifted content"""
        channel_id = random.randint(1, 500)
        start_time = (datetime.now() - timedelta(hours=random.randint(1, 24))).isoformat()

        self.client.get(
            f"/api/v1/iptv/dvr/channel_{channel_id}?start={start_time}&duration=3600",
            headers={"X-Stream-Token": self.stream_token},
        )

class IoTDeviceSimulator(HttpUser):
    """Simulates IoT device telemetry flood"""
    wait_time = between(0.1, 2)
    weight = 100

    def on_start(self):
        self.property_id = random.choice(["prop_alpha", "prop_beta", "prop_gamma"])
        self.device_id = f"device_{random.randint(10000, 99999)}"
        self.device_type = random.choice([
            "thermostat", "occupancy_sensor", "light", "lock",
            "curtain", "tv", "minibar", "safe", "phone",
        ])

    @task(50)
    def send_telemetry(self):
        """Send device telemetry data"""
        telemetry = {
            "device_id": self.device_id,
            "property_id": self.property_id,
            "timestamp": datetime.utcnow().isoformat(),
            "type": self.device_type,
            "data": self.generate_telemetry(),
        }

        self.client.post("/api/v1/iot/telemetry", json=telemetry)

    @task(10)
    def send_batch_telemetry(self):
        """Send batched telemetry (bulk ingestion)"""
        batch = [
            {
                "device_id": f"device_{random.randint(10000, 99999)}",
                "property_id": self.property_id,
                "timestamp": datetime.utcnow().isoformat(),
                "type": random.choice(["thermostat", "occupancy_sensor", "light"]),
                "data": self.generate_telemetry(),
            }
            for _ in range(100)
        ]

        self.client.post("/api/v1/iot/telemetry/batch", json={"devices": batch})

    @task(5)
    def trigger_alert(self):
        """Simulate IoT alert conditions"""
        alerts = [
            {"type": "motion_detected", "room_id": f"room_{random.randint(100, 9999)}"},
            {"type": "door_opened", "room_id": f"room_{random.randint(100, 9999)}"},
            {"type": "temperature_anomaly", "value": random.uniform(30, 40)},
            {"type": "low_battery", "device_id": self.device_id},
            {"type": "water_leak", "room_id": f"room_{random.randint(100, 9999)}"},
        ]

        self.client.post("/api/v1/iot/alerts", json=random.choice(alerts))

    def generate_telemetry(self):
        """Generate realistic device telemetry"""
        if self.device_type == "thermostat":
            return {
                "temperature": round(random.uniform(18, 26), 1),
                "humidity": round(random.uniform(30, 70), 1),
                "setpoint": random.randint(18, 26),
                "mode": random.choice(["cool", "heat", "auto", "off"]),
                "fan_speed": random.choice(["low", "medium", "high", "auto"]),
            }
        elif self.device_type == "occupancy_sensor":
            return {
                "occupied": random.choice([True, False]),
                "confidence": round(random.uniform(0.8, 1.0), 2),
                "motion_count": random.randint(0, 50),
            }
        elif self.device_type == "light":
            return {
                "on": random.choice([True, False]),
                "brightness": random.randint(0, 100),
                "color_temp": random.choice([2700, 3000, 4000, 5000, 6500]),
                "power_consumption": round(random.uniform(0, 60), 2),
            }
        elif self.device_type == "lock":
            return {
                "locked": random.choice([True, False]),
                "battery_level": random.randint(20, 100),
                "last_access": datetime.utcnow().isoformat(),
                "access_method": random.choice(["keycard", "mobile", "manual"]),
            }
        else:
            return {"status": "online", "timestamp": datetime.utcnow().isoformat()}

class AIInferenceUser(HttpUser):
    """Simulates AI inference load"""
    wait_time = between(0.5, 5)
    weight = 30

    def on_start(self):
        self.tenant_id = random.choice(["hotel-alpha", "hotel-beta"])

    @task(10)
    def concierge_query(self):
        """AI concierge inference"""
        queries = [
            "What restaurants are open now?",
            "How do I get to the airport?",
            "What's the weather tomorrow?",
            "Book a massage for 3 PM",
            "Recommend a local tour",
            "How do I use the TV?",
            "Order breakfast to my room",
            "What's the checkout time?",
        ]

        self.client.post("/api/v1/ai/concierge", json={
            "prompt": random.choice(queries),
            "tenant_id": self.tenant_id,
            "use_rag": True,
            "max_tokens": 150,
        })

    @task(5)
    def sentiment_analysis(self):
        """Guest sentiment analysis"""
        reviews = [
            "The room was amazing and the staff were very helpful!",
            "Terrible experience, the AC didn't work and nobody fixed it.",
            "Good location but the breakfast was mediocre.",
            "Absolutely loved the spa and the pool area.",
            "Room was clean but too small for the price.",
        ]

        self.client.post("/api/v1/ai/sentiment", json={
            "text": random.choice(reviews),
            "tenant_id": self.tenant_id,
        })

    @task(3)
    def occupancy_forecast(self):
        """Demand forecasting inference"""
        self.client.post("/api/v1/ai/forecast/occupancy", json={
            "property_id": f"prop_{random.randint(1, 10)}",
            "horizon_days": 30,
            "tenant_id": self.tenant_id,
        })

    @task(2)
    def recommendation_engine(self):
        """Personalized recommendations"""
        self.client.post("/api/v1/ai/recommendations", json={
            "guest_id": f"guest_{random.randint(1000, 9999)}",
            "context": {
                "time_of_day": random.choice(["morning", "afternoon", "evening", "night"]),
                "weather": random.choice(["sunny", "rainy", "cloudy"]),
                "previous_bookings": random.randint(0, 10),
            },
            "tenant_id": self.tenant_id,
        })

# Event hooks for metrics collection
@events.request.add_listener
def on_request(request_type, name, response_time, response_length, response, context, exception, **kwargs):
    if exception:
        # Log to Prometheus/Grafana
        print(f"REQUEST_FAILED: {name} | {exception}")

@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    if isinstance(environment.runner, MasterRunner):
        print("Load test completed. Generating report...")
        # Trigger report generation
        # Upload metrics to observability platform
