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
  const name = 'time-picker ' + c.string({ length: 10, alpha: true })
  await page.goto(`${baseURL}/schedules`)
  await page.getByRole('button', { name: 'Create Schedule' }).click()
  await page.fill('input[name=name]', name)
  await page.locator('button[type=submit]').click()
  await page.waitForURL(/\/schedules\/.{36}/)
  scheduleID = page.url().split('/schedules/')[1]
})

test.afterEach(async ({ page }) => {
  // delete schedule
  await page.goto(`${baseURL}/schedules/${scheduleID}`)
  await page.click('[data-testid="DeleteIcon"]')
  await page.click('button:has-text("Confirm")')
})

test('should handle selecting date values', async ({ page }) => {
  await page.goto(
    `${baseURL}/schedules/${scheduleID}/shifts?start=2006-01-02T06%3A00%3A00.000Z`,
  )
  await expect(page.locator('text=1/2/2006')).toContainText('1/2/2006')

  await page.click('button[title="Filter"]')
  await page.fill('input[name="filterStart"]', '2007-02-03')
  await page.locator('text=Done').click()
  await expect(page.locator('text=2/3/2007')).toContainText('2/3/2007')
})
