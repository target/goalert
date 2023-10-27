import { test, expect } from '@playwright/test'
import { adminSessionFile } from './lib'
import Chance from 'chance'
import { createService } from './lib/service'
const c = new Chance()
test.describe.configure({ mode: 'parallel' })
test.use({ storageState: adminSessionFile })

test('Admin', async ({ page }) => {
  const name = 'pw-service ' + c.name()
  const description = c.sentence()
  await createService(page, name, description)

  await page.goto('./admin/service-metrics')

  // wait for services to finish loading
  await expect(page.locator('.MuiDataGrid-overlay')).toHaveCount(0)

  // get totalServices count
  const totalServices = await page
    .locator('.MuiCardHeader-title')
    .nth(0)
    .innerText()
  expect(parseInt(totalServices)).toBeGreaterThan(0)

  // get services missing integrations count
  const missingInts = await page
    .locator('.MuiCardHeader-title')
    .nth(1)
    .innerText()
  expect(parseInt(missingInts)).toBeGreaterThan(0)

  // get services missing notifications count
  const missingNotifs = await page
    .locator('.MuiCardHeader-title')
    .nth(2)
    .innerText()
  expect(missingNotifs).toBe('0')

  // get services reaching alert limit
  const alertLimit = await page
    .locator('.MuiCardHeader-title')
    .nth(3)
    .innerText()
  expect(alertLimit).toBe('0')
})
