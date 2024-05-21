import { test, expect } from '@playwright/test'
import {
  baseURLFromFlags,
  dropdownSelect,
  pageAction,
  userSessionFile,
} from './lib'
import Chance from 'chance'
import { createService } from './lib/service'
const c = new Chance()

test.describe.configure({ mode: 'parallel' })
test.use({
  storageState: userSessionFile,
  baseURL: baseURLFromFlags(['univ-keys']),
})

test('Integration Keys', async ({ page, playwright, isMobile }) => {
  const intKeyName = 'uik-key ' + c.name()
  const serviceName = 'uik-service ' + c.name()
  const serviceDescription = c.sentence()
  const serviceURL = await createService(page, serviceName, serviceDescription)

  await page.getByRole('link', { name: 'Integration Keys' }).click()

  await pageAction(page, 'Create Integration Key')
  await page.fill('input[name=name]', intKeyName)
  await dropdownSelect(page, 'Type', 'Universal Integration Key')
  await page.click('button[type=submit]')
  await page.waitForURL(/\/services\/.+\/integration-keys\/.+$/)

  const bread = page.locator('header nav')
  await expect(bread.getByRole('link', { name: 'Services' })).toBeVisible()
  if (!isMobile) {
    // mobile collapses these
    await expect(bread.getByRole('link', { name: serviceName })).toBeVisible()
    await expect(
      bread.getByRole('link', { name: 'Integration Keys' }),
    ).toBeVisible()
  }
  await expect(bread.getByRole('link', { name: intKeyName })).toBeVisible()

  // validate token functionality
  await expect(page.locator('[data-cy=details]')).toContainText(
    'Auth Token: N/A',
  )

  // create first token
  await page.getByRole('button', { name: 'Generate Auth Token' }).click()
  await page.getByRole('button', { name: 'Generate' }).click()

  const origPrimaryToken = await page
    .getByRole('button', { name: 'Copy' })
    .textContent()
  await expect(origPrimaryToken).not.toBeNull()
  await page.getByRole('button', { name: 'Done' }).click()

  function tokenHint(token: string): string {
    return token.slice(0, 2) + '...' + token.slice(-4)
  }

  await expect(page.locator('[data-cy=details]')).toContainText(
    tokenHint(origPrimaryToken as string),
  )

  const req = await playwright.request.newContext({
    baseURL: baseURLFromFlags(['univ-keys']),
    extraHTTPHeaders: {
      'Content-Type': 'application/json',
      Cookie: '', // ensure no cookies are sent
    },
  })
  async function testKey(key: string, status: number): Promise<void> {
    const auth = { Authorization: '' }
    if (key) auth.Authorization = 'Bearer ' + key
    const resp = await req.post('/api/v2/uik', {
      headers: auth,
      data: {},
    })
    await expect(resp.status()).toBe(status)
  }

  await testKey('', 401)
  await testKey(origPrimaryToken as string, 204)

  await page.getByRole('button', { name: 'Generate Secondary Token' }).click()
  await page.getByRole('button', { name: 'Generate' }).click()

  const firstSecondaryToken = await page
    .getByRole('button', { name: 'Copy' })
    .textContent()
  await expect(origPrimaryToken).not.toBeNull()
  await page.getByRole('button', { name: 'Done' }).click()

  await expect(page.locator('[data-cy=details]')).toContainText(
    tokenHint(origPrimaryToken as string),
  )
  await expect(page.locator('[data-cy=details]')).toContainText(
    tokenHint(firstSecondaryToken as string),
  )

  // ensure both work
  await testKey(origPrimaryToken as string, 204)
  await testKey(firstSecondaryToken as string, 204)

  await page.getByRole('button', { name: 'Delete Secondary Token' }).click()
  await page.getByLabel('I acknowledge the impact of this action').click()
  await page.getByRole('button', { name: 'Submit' }).click()
  await expect(page.locator('[data-cy=details]')).not.toContainText(
    tokenHint(firstSecondaryToken as string),
  )
  await testKey(origPrimaryToken as string, 204)
  await testKey(firstSecondaryToken as string, 401)

  await page.getByRole('button', { name: 'Generate Secondary Token' }).click()
  await page.getByRole('button', { name: 'Generate' }).click()
  const secondSecondaryToken = await page
    .getByRole('button', { name: 'Copy' })
    .textContent()
  await expect(secondSecondaryToken).not.toBeNull()
  await page.getByRole('button', { name: 'Done' }).click()

  await page.getByRole('button', { name: 'Promote Secondary Token' }).click()
  await page.getByLabel('I acknowledge the impact of this action').click()
  await page.getByRole('button', { name: 'Promote Key' }).click()

  await expect(page.locator('[data-cy=details]')).toContainText(
    tokenHint(secondSecondaryToken as string),
  )
  await expect(page.locator('[data-cy=details]')).not.toContainText(
    tokenHint(origPrimaryToken as string),
  )

  await testKey(origPrimaryToken as string, 401)
  await testKey(secondSecondaryToken as string, 204)

  await page.goto(serviceURL)
  await page.getByRole('button', { name: 'Delete' }).click()
  await page.getByRole('button', { name: 'Confirm' }).click()
  await page.waitForURL(/services$/)
})
