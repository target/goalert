import { test, expect } from '@playwright/test'
import { userSessionFile } from './lib'
import Chance from 'chance'
const c = new Chance()
const baseURL = 'http://localhost:6130'

test.describe.configure({ mode: 'parallel' })
test.use({ storageState: userSessionFile })

let scheduleID: string

test.beforeEach(async ({ page }) => {
  // create schedule
  const name = c.name() + ' Service'
  await page.goto(`${baseURL}/schedules`)
  await page.getByRole('button', { name: 'Create Schedule' }).click()
  await page.fill('input[name=name]', name)
  await page.locator('button[type=submit]').click()
  await page.waitForTimeout(1000)
  const p = page.url().split('/')
  scheduleID = p[p.length - 1]
})

test.afterEach(async ({ page }) => {
  // delete schedule
  await page.goto(`${baseURL}/schedules/${scheduleID}`)
  await page.click('[data-testid="DeleteIcon"]')
  await page.click('button:has-text("Confirm")')
})

test('local time hover', async ({ page, isMobile }) => {
  // change schedule tz to Europe/Amsterdam
  await page.click('[aria-label="Edit"]')
  await page.fill('input[name=time-zone]', 'Europe/Amsterdam')
  await page.waitForTimeout(2000)
  for (let i = 0; i < 2; i++) await page.keyboard.press('Enter')

  // add user override
  await page.goto(`${baseURL}/schedules/${scheduleID}/shifts`)
  if (!isMobile) {
    await page.click('button:has-text("Create Override")')
    await page.keyboard.press('Tab')
    for (let i = 0; i < 2; i++) await page.keyboard.press('ArrowDown')
    await page.keyboard.press('Enter')
  } else {
    await page.click('[data-testid="AddIcon"]')
    for (let i = 0; i < 2; i++) await page.keyboard.press('Tab')
    for (let i = 0; i < 2; i++) await page.keyboard.press('ArrowDown')
    await page.keyboard.press('Enter')
  }

  // should display correct timezone in form
  await expect(page.locator('form[id=dialog-form]')).toContainText(
    'Times shown in schedule timezone (Europe/Amsterdam)',
  )

  await page.locator('input[name=addUserID]').fill('Admin McIntegrationFace')
  await page.waitForTimeout(1000)
  await page.getByText('Admin McIntegrationFace').click()
  await page.locator('button[type=submit]').click()

  // should display local tz on hover
  await page.hover('span:has-text("GMT")')
  await expect(page.locator('div[role=tooltip]')).not.toContainText('GMT')
})
