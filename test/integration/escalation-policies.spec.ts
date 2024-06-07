import { test, expect } from '@playwright/test'
import { userSessionFile } from './lib'
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
  await page.goto(`/rotations`)
  await page.getByRole('button', { name: 'Create Rotation' }).click()
  await page.fill('input[name=name]', rotName)
  await page.fill('textarea[name=description]', 'test rotation')
  await page.locator('button[type=submit]').click()
  await page.waitForURL(/\/rotations\/.{36}/)
  rotID = page.url().split('/rotations/')[1]

  // create schedule
  schedName = 'sched-' + c.string({ length: 10, alpha: true })
  await page.goto(`/schedules`)
  await page.getByRole('button', { name: 'Create Schedule' }).click()
  await page.fill('input[name=name]', schedName)
  await page.locator('button[type=submit]').click()
  await page.waitForURL(/\/schedules\/.{36}/)
  schedID = page.url().split('/schedules/')[1]

  // create EP
  epName = 'ep-' + c.string({ length: 10, alpha: true })
  await page.goto(`/escalation-policies`)
  await page.getByRole('button', { name: 'Create Escalation Policy' }).click()
  await page.fill('input[name=name]', epName)
  await page.locator('button[type=submit]').click()
  await page.waitForURL(/\/escalation-policies\/.{36}/)
  epID = page.url().split('/escalation-policies/')[1]
})

test.afterEach(async ({ page }) => {
  // delete rotation
  await page.goto(`/rotations/${rotID}`)
  await page.click('[data-testid="DeleteIcon"]')
  await page.click('button:has-text("Confirm")')

  // delete schedule
  await page.goto(`/schedules/${schedID}`)
  await page.click('[data-testid="DeleteIcon"]')
  await page.click('button:has-text("Confirm")')

  // delete EP
  await page.goto(`/escalation-policies/${epID}`)
  await page.click('[data-testid="DeleteIcon"]')
  await page.click('button:has-text("Confirm")')
})

test('check caption delay text while creating steps', async ({ page }) => {
  await page.goto(`/escalation-policies/${epID}`)

  async function createStep(delayMinutes: string) {
    await page.getByRole('button', { name: 'Create Step' }).click()
    await page.getByLabel('Destination Type').click()
    await page.locator('li', { hasText: 'User' }).click()
    await page.getByRole('combobox', { name: 'User', exact: true }).click()
    await page
      .getByRole('combobox', { name: 'User', exact: true })
      .fill('Admin McIntegrationFace')
    await page
      .locator('div[role=presentation]')
      .locator('li', { hasText: 'Admin McIntegrationFace' })
      .click()
    await page.getByRole('button', { name: 'Add Destination' }).click()
    await expect(
      page
        .getByRole('dialog')
        .getByTestId('destination-chip')
        .filter({ hasText: 'Admin McIntegrationFace' }),
    ).toBeVisible()
    await page.locator('input[name=delayMinutes]').fill(delayMinutes)
    await page.locator('button[type=submit]', { hasText: 'Submit' }).click()
  }

  await createStep('5')
  await expect(
    page.getByText('Go back to step #1 after 5 minutes'),
  ).toBeVisible()
  await createStep('20')
  await expect(
    page.getByText('Move on to step #2 after 5 minutes'),
  ).toBeVisible()
  await expect(
    page.getByText('Go back to step #1 after 20 minutes'),
  ).toBeVisible()
})

test('create escalation policy step using destination actions', async ({
  page,
}) => {
  await page.goto(`/escalation-policies/${epID}`)
  await page.getByRole('button', { name: 'Create Step' }).click()

  // add rotation
  await page.getByLabel('Destination Type').click()
  await page.locator('li', { hasText: 'Rotation' }).click()
  await page.getByRole('combobox', { name: 'Rotation', exact: true }).click()
  await page
    .getByRole('combobox', { name: 'Rotation', exact: true })
    .fill(rotName)
  await page.locator('li', { hasText: rotName }).click()
  await page.getByRole('button', { name: 'Add Destination' }).click()
  await expect(
    page
      .getByRole('dialog')
      .getByTestId('destination-chip')
      .filter({ hasText: rotName }),
  ).toBeVisible()

  // add schedule
  await page.getByLabel('Destination Type').click()
  await page.locator('li', { hasText: 'Schedule' }).click()
  await page.getByRole('combobox', { name: 'Schedule', exact: true }).click()
  await page
    .getByRole('combobox', { name: 'Schedule', exact: true })
    .fill(schedName)
  await page.locator('li', { hasText: schedName }).click()
  await page.getByRole('button', { name: 'Add Destination' }).click()
  await expect(
    page
      .getByRole('dialog')
      .getByTestId('destination-chip')
      .filter({ hasText: schedName }),
  ).toBeVisible()

  // add user
  await page.getByLabel('Destination Type').click()
  await page.locator('li', { hasText: 'User' }).click()
  await page.getByRole('combobox', { name: 'User', exact: true }).click()
  await page
    .getByRole('combobox', { name: 'User', exact: true })
    .fill('Admin McIntegrationFace')
  await page.locator('li', { hasText: 'Admin McIntegrationFace' }).click()
  await page.getByRole('button', { name: 'Add Destination' }).click()
  await expect(
    page
      .getByRole('dialog')
      .getByTestId('destination-chip')
      .filter({ hasText: 'Admin McIntegrationFace' }),
  ).toBeVisible()

  await page.locator('button[type=submit]', { hasText: 'Submit' }).click()

  // expect to see new step information on ep page
  await expect(page.getByText('Step #1:')).toBeVisible()

  const rotLink = await page.locator('a', { hasText: rotName })
  await expect(rotLink).toHaveAttribute(
    'href',
    new RegExp(`/rotations/${rotID}$`),
  )

  const schedLink = await page.locator('a', { hasText: schedName })
  await expect(schedLink).toHaveAttribute(
    'href',
    new RegExp(`/schedules/${schedID}$`),
  )

  await expect(
    await page.locator('a', {
      hasText: 'Admin McIntegrationFace',
    }),
  ).toBeVisible()
})
