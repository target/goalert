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
  await page.click('input[name=time-zone]')
  await page.keyboard.type('Europe/Amsterdam')
  await page.waitForTimeout(500)
  await page.keyboard.press('ArrowDown')
  await page.keyboard.press('Enter')
  await page.locator('button[type=submit]').click()
  await page.waitForTimeout(500)
  const p = page.url().split('/')
  scheduleID = p[p.length - 1]
})

test.afterEach(async ({ page }) => {
  // delete schedule
  await page.goto(`${baseURL}/schedules/${scheduleID}`)
  await page.click('[data-testid="DeleteIcon"]')
  await page.click('button:has-text("Confirm")')
})

test.only('local time hover', async ({ page }) => {
  await page.click('span:has-text("Shifts")')
  await page.hover('[data-testid="AddIcon"]')
  await page.click('[data-testid="AccountPlusIcon"]')

  // should display correct timezone in form
  await expect(page.locator('form[id=dialog-form]')).toContainText(
    'Times shown in schedule timezone (Europe/Amsterdam)',
  )

  // add user override
  await page.click('input[name=addUserID]')
  await page.waitForTimeout(500)
  await page.keyboard.press('ArrowDown')
  await page.waitForTimeout(500)
  await page.keyboard.press('Enter')
  await page.waitForTimeout(500)
  for (let i = 0; i < 4; i++) await page.keyboard.press('Tab')
  await page.keyboard.type('1')
  await page.keyboard.press('Tab')
  await page.keyboard.type('0')
  await page.keyboard.press('Tab')
  await page.keyboard.type('p')
  for (let i = 0; i < 2; i++) await page.keyboard.press('Tab')
  await page.keyboard.press('ArrowUp')
  for (let i = 0; i < 2; i++) await page.keyboard.press('Tab')
  await page.keyboard.type('5')
  await page.keyboard.press('Tab')
  await page.keyboard.type('0')
  await page.keyboard.press('Tab')
  await page.keyboard.type('a')
  for (let i = 0; i < 3; i++) await page.keyboard.press('Tab')
  await page.keyboard.press('Enter')

  // should display local tz on hover
  await page.goto(`${baseURL}/schedules/${scheduleID}/overrides`)
  await page.hover('span:has-text("5:00 AM GMT+2")')
  await expect(page.locator('div[role=tooltip]')).toContainText('10:00 PM CDT')
})
