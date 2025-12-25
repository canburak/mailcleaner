import { test, expect } from '@playwright/test'

test.describe('MailCleaner E2E Tests', () => {
  test('should load the home page', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveTitle(/MailCleaner/i)
  })

  test('should navigate to accounts page', async ({ page }) => {
    await page.goto('/')
    // Add navigation tests based on your app structure
  })

  test('should display accounts list', async ({ page }) => {
    await page.goto('/accounts')
    // Verify accounts list renders
  })
})
