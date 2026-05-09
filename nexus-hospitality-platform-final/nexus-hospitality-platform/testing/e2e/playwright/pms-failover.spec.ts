# testing/e2e/playwright/pms-failover.spec.ts
import { test, expect } from '@playwright/test';
import { PMSPage } from './pages/PMSPage';

test.describe('PMS Failover Scenarios', () => {
  let pmsPage: PMSPage;

  test.beforeEach(async ({ page }) => {
    pmsPage = new PMSPage(page);
    await pmsPage.goto();
    await pmsPage.login('admin', 'password');
  });

  test('should handle database failover gracefully', async ({ page }) => {
    // Simulate database failover
    await page.evaluate(() => {
      window.localStorage.setItem('simulate_db_failover', 'true');
    });

    // Attempt reservation creation
    await pmsPage.createReservation({
      guestName: 'Test Guest',
      checkIn: '2024-06-01',
      checkOut: '2024-06-05',
    });

    // Verify graceful degradation
    await expect(page.locator('[data-testid="failover-banner"]')).toBeVisible();
    await expect(page.locator('[data-testid="reservation-queued"]')).toBeVisible();
  });

  test('should maintain session during regional outage', async ({ page }) => {
    // Simulate regional outage
    await page.route('**/api/v1/**', async (route) => {
      if (route.request().url().includes('us-east-1')) {
        await route.abort('timedout');
      } else {
        await route.continue();
      }
    });

    // Verify automatic failover to secondary region
    await pmsPage.navigateToDashboard();
    await expect(page.locator('[data-testid="region-indicator"]')).toContainText('us-west-2');
    await expect(page.locator('[data-testid="dashboard-content"]')).toBeVisible();
  });

  test('should handle Kafka outage with queue persistence', async ({ page }) => {
    // Create event while Kafka is down
    await pmsPage.triggerKafkaOutage();

    const reservationId = await pmsPage.createReservation({
      guestName: 'Kafka Test',
      checkIn: '2024-06-10',
      checkOut: '2024-06-15',
    });

    // Verify event queued locally
    await expect(page.locator('[data-testid="event-queue-count"]')).toHaveText('1');

    // Restore Kafka and verify replay
    await pmsPage.restoreKafka();
    await expect(page.locator('[data-testid="event-queue-count"]')).toHaveText('0');

    // Verify reservation propagated
    await pmsPage.searchReservation(reservationId);
    await expect(page.locator('[data-testid="reservation-found"]')).toBeVisible();
  });
});
