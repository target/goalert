import { test, expect } from '@playwright/test'
import { baseURLFromFlags, userSessionFile } from './lib'
import Chance from 'chance'
const c = new Chance()

test.use({ storageState: userSessionFile })

let rotID: string
let rotName: string

let schedID: string
let schedName: string

let epID: string
let epName: string

test.beforeEach(async ({ page }) => {
  // create rotation
  rotName = 'rot-' + c.string({ length: 10, alpha: true })
  await page.goto(`${baseURLFromFlags(['dest-types'])}/rotations`)
  await page.getByRole('button', { name: 'Create Rotation' }).click()
  await page.fill('input[name=name]', rotName)
  await page.fill('textarea[name=description]', 'test rotation')
  await page.locator('button[type=submit]').click()
  await page.waitForURL(/\/rotations\/.{36}/)
  rotID = page.url().split('/rotations/')[1]

  // create schedule
  schedName = 'sched-' + c.string({ length: 10, alpha: true })
  await page.goto(`${baseURLFromFlags(['dest-types'])}/schedules`)
  await page.getByRole('button', { name: 'Create Schedule' }).click()
  await page.fill('input[name=name]', schedName)
  await page.locator('button[type=submit]').click()
  await page.waitForURL(/\/schedules\/.{36}/)
  schedID = page.url().split('/schedules/')[1]

  // create EP
  epName = 'ep-' + c.string({ length: 10, alpha: true })
  await page.goto(`${baseURLFromFlags(['dest-types'])}/escalation-policies`)
  await page.getByRole('button', { name: 'Create Escalation Policy' }).click()
  await page.fill('input[name=name]', epName)
  await page.locator('button[type=submit]').click()
  await page.waitForURL(/\/escalation-policies\/.{36}/)
  epID = page.url().split('/escalation-policies/')[1]
})

test.afterEach(async ({ page }) => {
  // delete rotation
  await page.goto(`${baseURLFromFlags(['dest-types'])}/rotations/${rotID}`)
  await page.click('[data-testid="DeleteIcon"]')
  await page.click('button:has-text("Confirm")')

  // delete schedule
  await page.goto(`${baseURLFromFlags(['dest-types'])}/schedules/${schedID}`)
  await page.click('[data-testid="DeleteIcon"]')
  await page.click('button:has-text("Confirm")')

  // delete EP
  await page.goto(
    `${baseURLFromFlags(['dest-types'])}/escalation-policies/${epID}`,
  )
  await page.click('[data-testid="DeleteIcon"]')
  await page.click('button:has-text("Confirm")')
})

test('create escalation policy step using destination actions', async ({
  page,
}) => {
  await page.goto(
    `${baseURLFromFlags(['dest-types'])}/escalation-policies/${epID}`,
  )
  await page.getByRole('button', { name: 'Create Step' }).click()

  // add rotation
  await page.getByPlaceholder('Start typing...').click()
  await page.keyboard.type(rotName)
  await page.click(`li:has-text("${rotName}")`)
  await page.getByRole('button', { name: 'Add Action' }).click()

  // expect to see rotation added
  await expect(page.getByText(rotName)).toBeVisible()
  await expect(page.getByTestId('RotateRightIcon')).toHaveCount(2)

  // add schedule
  await page.locator('text=Rotation >> nth=1').click()
  await page.keyboard.press('ArrowDown')
  await page.keyboard.press('Enter')
  await page.getByPlaceholder('Start typing...').click()
  await page.keyboard.type(schedName)
  await page.click(`li:has-text("${schedName}")`)
  await page.getByRole('button', { name: 'Add Action' }).click()

  // expect to see schedule added
  await expect(page.getByText(schedName)).toBeVisible()
  await expect(page.getByTestId('TodayIcon')).toHaveCount(2)

  // add user
  await page.locator('text=Schedule >> nth=1').click()
  await page.keyboard.press('ArrowDown')
  await page.keyboard.press('Enter')
  await page.getByPlaceholder('Start typing...').click()
  await page.keyboard.type('Admin')
  await page.click('li:has-text("Admin McIntegrationFace")')
  await page.getByRole('button', { name: 'Add Action' }).click()
  await expect(page.getByTestId('spinner')).toBeHidden()

  // expect to see user name added
  await expect(page.getByText('Admin McIntegrationFace')).toBeVisible()

  // expect to see new step information on ep page
  await page.locator('button[type=submit] >> nth=1').click()
  await expect(page.getByText('Step #1:')).toBeVisible()
  await expect(page.getByText(rotName)).toBeVisible()
  await expect(page.getByTestId('RotateRightIcon')).toHaveCount(2)
  await expect(page.getByText(schedName)).toBeVisible()
  await expect(page.getByTestId('TodayIcon')).toHaveCount(2)
  await expect(page.getByText('Admin McIntegrationFace')).toBeVisible()
})
