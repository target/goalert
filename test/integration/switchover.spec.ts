import { test, expect } from '@playwright/test'
import { adminUserCreds } from './lib'
const baseURL = 'http://localhost:6110'

test('should perform switchover', async ({ page }) => {
  await page.goto(`${baseURL}/admin/switchover`)
  await page.fill('input[name=username]', adminUserCreds.user)
  await page.fill('input[name=password]', adminUserCreds.pass)
  await page.click('button[type=submit] >> "Login"')
  await expect(page.locator('main')).toContainText('Switchover Status')
  await expect(page.locator('main')).toContainText('Needs Reset')

  await page.click('button >> "Reset"')

  // will take some time to reset
  await expect(page.locator('main')).toContainText('Ready', { timeout: 30000 })

  await page.click('button >> "Execute"')

  // will take some time to execute
  await expect(page.locator('main')).toContainText(
    'DB switchover is complete',
    { timeout: 90000 },
  )
})
