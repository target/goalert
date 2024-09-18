import { test, expect } from '@playwright/test'
import { dropdownSelect, userSessionFile } from './lib'
import Chance from 'chance'
const c = new Chance()

test.describe.configure({ mode: 'parallel' })
test.use({ storageState: userSessionFile })

// test create, edit, verify, and delete of an EMAIL contact method
test('first time setup', async ({ page }) => {
  await page.goto('.?isFirstLogin=1')
  // ensure dialog is shown
  await expect(
    page.locator('[role=dialog]', { hasText: 'Welcome to GoAlert' }),
  ).toBeVisible()

  const name = 'first-setup-email ' + c.name()
  const email = 'first-setup-email-' + c.email()
  await page.fill('input[name=name]', name)
  await dropdownSelect(page, 'Destination Type', 'Email')
  await page.fill('input[name=email_address]', email)
  await page.click('[role=dialog] button[type=submit]')
  await expect(
    page.locator('[role=dialog]', { hasText: 'Verify Contact Method' }),
  ).toBeVisible()

  // cancel out
  await page.locator('[role=dialog] button', { hasText: 'Cancel' }).click()

  // ensure dialog is not shown
  await expect(page.locator('[role=dialog]')).toBeHidden()

  await page.goto('./profile')
  await page.click(`li:has-text("${email}") [aria-label="Other Actions"]`)

  await page.getByRole('menuitem', { name: 'Delete' }).click()
  await page.getByRole('button', { name: 'Confirm' }).click()

  await expect(page.locator('[role=dialog]')).toBeHidden()
  await expect(
    page.locator(`li:has-text("${email}") [aria-label="Other Actions"]`),
  ).toBeHidden()
})
