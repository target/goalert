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
  await page.click('[aria-label="Create Schedule"]')
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

test('local time hover', async ({ page }) => {
  // change schedule tz to Europe/Amsterdam
  await page.click('[aria-label="Edit"]')
  await page.fill('input[name=time-zone]', 'Europe/Amsterdam')
  await page.waitForTimeout(2000)
  await page.keyboard.press('ArrowDown')
  for (let i = 0; i < 2; i++) await page.keyboard.press('Enter')

  // add user override
  await page.click('span:has-text("Shifts")')
  await page.hover('[data-testid="AddIcon"]')
  await page.click('[data-testid="AccountPlusIcon"]')

  // should display correct timezone in form
  await expect(page.locator('form[id=dialog-form]')).toContainText(
    'Times shown in schedule timezone (Europe/Amsterdam)',
  )

  await page.click('input[name=addUserID]')
  await page.waitForTimeout(1000)
  await page.keyboard.press('ArrowDown')
  await page.waitForTimeout(1000)
  await page.keyboard.press('Enter')
  await page.waitForTimeout(1000)
  await page.locator('button[type=submit]').click()

  // should display local tz on hover
  await page.hover('span:has-text("GMT+2")')
  await expect(page.locator('div[role=tooltip]')).toContainText('CDT')
})
