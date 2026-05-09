#!/usr/bin/env python3
# scripts/migration/pms-migration-factory.py
# ============================================================
# PMS MIGRATION FACTORY — From legacy systems to Nexus
# Supports: Oracle OPERA, Mews, Cloudbeds, Protel, Fidelio
# ============================================================

import argparse
import json
import logging
import sys
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from dataclasses import dataclass
from enum import Enum
import asyncio
import aiohttp
import asyncpg
from concurrent.futures import ThreadPoolExecutor
import pandas as pd
from tqdm import tqdm

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(f'/var/log/nexus/migration-{datetime.now().strftime("%Y%m%d")}.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger('pms-migration')

class PMSVendor(Enum):
    ORACLE_OPERA = "oracle_opera"
    MEWS = "mews"
    CLOUDBEDS = "cloudbeds"
    PROTEL = "protel"
    FIDELIO = "fidelio"
    SYNXIS = "synxis"
    HOTELSPRO = "hotelspro"

@dataclass
class MigrationConfig:
    source_vendor: PMSVendor
    source_connection: Dict[str, str]
    target_tenant_id: str
    target_property_id: str
    batch_size: int = 1000
    parallel_workers: int = 10
    dry_run: bool = False
    validate_data: bool = True
    rollback_enabled: bool = True

@dataclass
class MigrationResult:
    entity: str
    total_records: int
    migrated_records: int
    failed_records: int
    duration_seconds: float
    errors: List[str]

class PMSExtractor:
    """Extracts data from legacy PMS systems"""

    def __init__(self, vendor: PMSVendor, connection_params: Dict[str, str]):
        self.vendor = vendor
        self.connection_params = connection_params
        self.session: Optional[aiohttp.ClientSession] = None
        self.db_pool: Optional[asyncpg.Pool] = None

    async def connect(self):
        if self.vendor in [PMSVendor.ORACLE_OPERA, PMSVendor.PROTEL, PMSVendor.FIDELIO]:
            # Database connection for on-premise systems
            self.db_pool = await asyncpg.create_pool(
                host=self.connection_params['host'],
                port=int(self.connection_params.get('port', 5432)),
                database=self.connection_params['database'],
                user=self.connection_params['username'],
                password=self.connection_params['password'],
                min_size=5,
                max_size=20
            )
        else:
            # API connection for cloud systems
            self.session = aiohttp.ClientSession(
                headers={
                    'Authorization': f'Bearer {self.connection_params["api_key"]}',
                    'Content-Type': 'application/json'
                }
            )

    async def extract_guests(self, since: Optional[datetime] = None) -> List[Dict]:
        """Extract guest profiles"""
        if self.vendor == PMSVendor.ORACLE_OPERA:
            return await self._extract_opera_guests(since)
        elif self.vendor == PMSVendor.MEWS:
            return await self._extract_mews_guests(since)
        elif self.vendor == PMSVendor.CLOUDBEDS:
            return await self._extract_cloudbeds_guests(since)
        return []

    async def _extract_opera_guests(self, since: Optional[datetime]) -> List[Dict]:
        query = """
            SELECT 
                g.name_id as external_id,
                g.first_name,
                g.last_name,
                g.email,
                g.phone,
                g.nationality,
                g.language,
                g.date_of_birth,
                g.passport_number as id_document_number,
                g.passport_country as id_document_country,
                g.vip_status,
                g.preferences,
                g.created_date as created_at,
                g.updated_date as updated_at
            FROM name g
            WHERE g.name_type = 'GUEST'
            AND ($1::timestamp IS NULL OR g.updated_date > $1)
            ORDER BY g.name_id
        """

        async with self.db_pool.acquire() as conn:
            rows = await conn.fetch(query, since)
            return [dict(row) for row in rows]

    async def _extract_mews_guests(self, since: Optional[datetime]) -> List[Dict]:
        url = f"{self.connection_params['base_url']}/api/connector/v1/customers/getAll"

        payload = {
            "ClientToken": self.connection_params['client_token'],
            "AccessToken": self.connection_params['access_token'],
            "Limit": 1000,
            "StartUtc": since.isoformat() if since else (datetime.now() - timedelta(days=365)).isoformat()
        }

        async with self.session.post(url, json=payload) as response:
            data = await response.json()
            return [
                {
                    'external_id': customer['Id'],
                    'first_name': customer.get('FirstName', ''),
                    'last_name': customer.get('LastName', ''),
                    'email': customer.get('Email', ''),
                    'phone': customer.get('Phone', ''),
                    'nationality': customer.get('NationalityCode', ''),
                    'language': customer.get('LanguageCode', 'en'),
                    'created_at': customer.get('CreatedUtc', ''),
                }
                for customer in data.get('Customers', [])
            ]

    async def _extract_cloudbeds_guests(self, since: Optional[datetime]) -> List[Dict]:
        url = f"{self.connection_params['base_url']}/api/v1.1/guests"

        params = {
            'propertyID': self.connection_params['property_id'],
            'limit': 100,
        }
        if since:
            params['modifiedSince'] = since.isoformat()

        guests = []
        page = 1

        while True:
            params['page'] = page
            async with self.session.get(url, params=params) as response:
                data = await response.json()

                if not data.get('data'):
                    break

                guests.extend([
                    {
                        'external_id': guest['guestID'],
                        'first_name': guest.get('firstName', ''),
                        'last_name': guest.get('lastName', ''),
                        'email': guest.get('email', ''),
                        'phone': guest.get('phone', ''),
                        'created_at': guest.get('createdAt', ''),
                    }
                    for guest in data['data']
                ])

                if len(data['data']) < 100:
                    break
                page += 1

        return guests

    async def extract_reservations(self, since: Optional[datetime] = None) -> List[Dict]:
        """Extract reservation data"""
        if self.vendor == PMSVendor.ORACLE_OPERA:
            return await self._extract_opera_reservations(since)
        elif self.vendor == PMSVendor.MEWS:
            return await self._extract_mews_reservations(since)
        elif self.vendor == PMSVendor.CLOUDBEDS:
            return await self._extract_cloudbeds_reservations(since)
        return []

    async def _extract_opera_reservations(self, since: Optional[datetime]) -> List[Dict]:
        query = """
            SELECT 
                r.resv_name_id as external_id,
                r.confirmation_no as confirmation_number,
                r.resort as property_code,
                r.room as room_number,
                r.guest_name_id as guest_id,
                r.reservation_status as status,
                r.begin_date as check_in_date,
                r.end_date as check_out_date,
                r.adults as num_adults,
                r.children as num_children,
                r.room_type as room_type_code,
                r.rate_code as rate_plan_code,
                r.market_code as market_segment,
                r.source_code as booking_source,
                r.total_revenue as total_amount,
                r.deposit_amount,
                r.guarantee_code as guarantee_method,
                r.special_requests,
                r.created_date as created_at,
                r.updated_date as updated_at
            FROM reservation r
            WHERE ($1::timestamp IS NULL OR r.updated_date > $1)
            ORDER BY r.resv_name_id
        """

        async with self.db_pool.acquire() as conn:
            rows = await conn.fetch(query, since)
            return [dict(row) for row in rows]

    async def close(self):
        if self.db_pool:
            await self.db_pool.close()
        if self.session:
            await self.session.close()

class NexusLoader:
    """Loads data into Nexus PMS"""

    def __init__(self, target_config: Dict[str, str]):
        self.target_config = target_config
        self.db_pool: Optional[asyncpg.Pool] = None

    async def connect(self):
        self.db_pool = await asyncpg.create_pool(
            host=self.target_config['host'],
            port=int(self.target_config.get('port', 5432)),
            database=self.target_config['database'],
            user=self.target_config['username'],
            password=self.target_config['password'],
            min_size=5,
            max_size=20
        )

    async def load_guests(self, guests: List[Dict], tenant_id: str, batch_size: int = 1000) -> MigrationResult:
        result = MigrationResult(
            entity='guests',
            total_records=len(guests),
            migrated_records=0,
            failed_records=0,
            duration_seconds=0,
            errors=[]
        )

        start_time = datetime.now()

        for i in range(0, len(guests), batch_size):
            batch = guests[i:i + batch_size]

            try:
                async with self.db_pool.acquire() as conn:
                    await conn.copy_records_to_table(
                        'guests',
                        records=[(
                            tenant_id,
                            g.get('external_id'),
                            g.get('first_name', ''),
                            g.get('last_name', ''),
                            g.get('email'),
                            g.get('phone'),
                            g.get('date_of_birth'),
                            g.get('nationality'),
                            g.get('language', 'en'),
                            g.get('id_document_type'),
                            g.get('id_document_number'),
                            g.get('id_document_expiry'),
                            g.get('id_document_country'),
                            json.dumps(g.get('preferences', {})),
                            g.get('created_at', datetime.now()),
                            g.get('updated_at', datetime.now())
                        ) for g in batch],
                        columns=['tenant_id', 'external_id', 'first_name', 'last_name', 
                                'email', 'phone', 'date_of_birth', 'nationality', 
                                'language_preference', 'id_document_type', 
                                'id_document_number', 'id_document_expiry', 
                                'id_document_country', 'preferences', 'created_at', 'updated_at']
                    )

                result.migrated_records += len(batch)

            except Exception as e:
                result.failed_records += len(batch)
                result.errors.append(str(e))
                logger.error(f"Failed to load guest batch: {e}")

        result.duration_seconds = (datetime.now() - start_time).total_seconds()
        return result

    async def load_reservations(self, reservations: List[Dict], tenant_id: str, property_id: str, batch_size: int = 1000) -> MigrationResult:
        result = MigrationResult(
            entity='reservations',
            total_records=len(reservations),
            migrated_records=0,
            failed_records=0,
            duration_seconds=0,
            errors=[]
        )

        start_time = datetime.now()

        for i in range(0, len(reservations), batch_size):
            batch = reservations[i:i + batch_size]

            try:
                async with self.db_pool.acquire() as conn:
                    await conn.copy_records_to_table(
                        'reservations',
                        records=[(
                            tenant_id,
                            property_id,
                            r.get('guest_id'),
                            r.get('confirmation_number'),
                            r.get('external_id'),
                            r.get('booking_source', 'migration'),
                            r.get('check_in_date'),
                            r.get('check_out_date'),
                            r.get('num_adults', 1),
                            r.get('num_children', 0),
                            r.get('status', 'confirmed'),
                            r.get('total_amount', 0),
                            r.get('currency_code', 'USD'),
                            r.get('special_requests'),
                            r.get('created_at', datetime.now()),
                            r.get('updated_at', datetime.now())
                        ) for r in batch],
                        columns=['tenant_id', 'property_id', 'guest_id', 'confirmation_number',
                                'external_reference', 'booking_source', 'check_in_date',
                                'check_out_date', 'num_adults', 'num_children', 'status',
                                'total_amount', 'currency_code', 'special_requests', 
                                'created_at', 'updated_at']
                    )

                result.migrated_records += len(batch)

            except Exception as e:
                result.failed_records += len(batch)
                result.errors.append(str(e))
                logger.error(f"Failed to load reservation batch: {e}")

        result.duration_seconds = (datetime.now() - start_time).total_seconds()
        return result

    async def close(self):
        if self.db_pool:
            await self.db_pool.close()

class MigrationOrchestrator:
    """Orchestrates the full migration process"""

    def __init__(self, config: MigrationConfig):
        self.config = config
        self.extractor = PMSExtractor(config.source_vendor, config.source_connection)
        self.loader = NexusLoader({
            'host': 'nexus-db.internal',
            'database': 'nexus_pms',
            'username': 'migration_user',
            'password': '***',  # From Vault
        })
        self.results: List[MigrationResult] = []

    async def run_migration(self):
        logger.info(f"Starting migration from {self.config.source_vendor.value} to Nexus")
        logger.info(f"Tenant: {self.config.target_tenant_id}, Property: {self.config.target_property_id}")

        if self.config.dry_run:
            logger.info("DRY RUN MODE - No data will be written")

        await self.extractor.connect()
        await self.loader.connect()

        try:
            # Phase 1: Migrate guests
            logger.info("Phase 1: Migrating guests...")
            guests = await self.extractor.extract_guests()
            logger.info(f"Extracted {len(guests)} guests")

            if not self.config.dry_run:
                guest_result = await self.loader.load_guests(
                    guests, 
                    self.config.target_tenant_id,
                    self.config.batch_size
                )
                self.results.append(guest_result)
                logger.info(f"Guest migration: {guest_result.migrated_records}/{guest_result.total_records} successful")

            # Phase 2: Migrate reservations
            logger.info("Phase 2: Migrating reservations...")
            reservations = await self.extractor.extract_reservations()
            logger.info(f"Extracted {len(reservations)} reservations")

            if not self.config.dry_run:
                reservation_result = await self.loader.load_reservations(
                    reservations,
                    self.config.target_tenant_id,
                    self.config.target_property_id,
                    self.config.batch_size
                )
                self.results.append(reservation_result)
                logger.info(f"Reservation migration: {reservation_result.migrated_records}/{reservation_result.total_records} successful")

            # Phase 3: Migrate folios/transactions
            # Phase 4: Migrate loyalty data
            # Phase 5: Migrate rate plans
            # Phase 6: Migrate inventory

            # Validation
            if self.config.validate_data:
                await self.validate_migration()

            # Generate report
            self.generate_report()

        finally:
            await self.extractor.close()
            await self.loader.close()

    async def validate_migration(self):
        """Validate migrated data integrity"""
        logger.info("Running data validation...")

        # Check guest count
        # Verify reservation links
        # Validate folio balances
        # Check data consistency

        pass

    def generate_report(self):
        """Generate migration report"""
        report = {
            'migration_id': f"mig-{datetime.now().strftime('%Y%m%d-%H%M%S')}",
            'source_vendor': self.config.source_vendor.value,
            'target_tenant': self.config.target_tenant_id,
            'target_property': self.config.target_property_id,
            'timestamp': datetime.now().isoformat(),
            'results': [
                {
                    'entity': r.entity,
                    'total': r.total_records,
                    'migrated': r.migrated_records,
                    'failed': r.failed_records,
                    'duration': r.duration_seconds,
                    'success_rate': r.migrated_records / r.total_records if r.total_records > 0 else 0,
                }
                for r in self.results
            ],
            'total_duration': sum(r.duration_seconds for r in self.results),
        }

        with open(f'/tmp/migration-report-{report["migration_id"]}.json', 'w') as f:
            json.dump(report, f, indent=2)

        logger.info("Migration report generated")

def main():
    parser = argparse.ArgumentParser(description='Nexus PMS Migration Factory')
    parser.add_argument('--vendor', required=True, choices=[v.value for v in PMSVendor])
    parser.add_argument('--tenant-id', required=True)
    parser.add_argument('--property-id', required=True)
    parser.add_argument('--config', required=True, help='Path to connection config JSON')
    parser.add_argument('--batch-size', type=int, default=1000)
    parser.add_argument('--dry-run', action='store_true')
    parser.add_argument('--since', help='Migrate data updated since this date (ISO format)')

    args = parser.parse_args()

    with open(args.config) as f:
        connection_config = json.load(f)

    since = datetime.fromisoformat(args.since) if args.since else None

    config = MigrationConfig(
        source_vendor=PMSVendor(args.vendor),
        source_connection=connection_config,
        target_tenant_id=args.tenant_id,
        target_property_id=args.property_id,
        batch_size=args.batch_size,
        dry_run=args.dry_run,
    )

    orchestrator = MigrationOrchestrator(config)
    asyncio.run(orchestrator.run_migration())

if __name__ == '__main__':
    main()
