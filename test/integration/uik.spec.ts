import { test, expect, Page } from '@playwright/test'
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

async function setup(page: Page, isMobile: boolean): Promise<string> {
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

  return serviceURL
}

test('create universal key, add rule with action', async ({
  page,
  isMobile,
}) => {
  await setup(page, isMobile)
  const ruleName = c.name()
  const ruleDesc = c.sentence({ words: 5 })
  const ruleNewDesc = c.sentence({ words: 3 })

  // create a rule
  await page.getByRole('button', { name: 'Create Rule' }).click()
  await page.fill('input[name=name]', ruleName)
  await page.fill('input[name=description]', ruleDesc)
  const editor = await page
    .getByTestId('code-conditionExpr')
    .locator('.cm-editor')
  await editor.click()
  await page.keyboard.insertText('true')
  await page.getByRole('button', { name: 'Next' }).click()

  // add an action to the rule and submit
  await dropdownSelect(page, 'Destination Type', 'Alert')
  await expect(
    page.locator('#dialog-form').getByTestId('no-actions'),
  ).toBeVisible()
  await page.getByRole('button', { name: 'Add Action' }).click()
  await expect(
    page.getByLabel('Create Rule').getByTestId('no-actions'),
  ).toBeHidden()
  await expect(
    page.locator('span', { hasText: 'Create new alert' }),
  ).toBeVisible()

  await page.getByRole('button', { name: 'Next' }).click()
  await page.getByRole('button', { name: 'Submit' }).click()
  await expect(page.locator('[role=dialog]')).toBeHidden()

  // verify
  await expect(page.locator('body')).toContainText(ruleName)
  await expect(page.locator('body')).toContainText(ruleDesc)

  // start editing
  await page.getByRole('button', { name: 'Other Actions' }).click()
  await page.getByRole('menuitem', { name: 'Edit' }).click()

  // update description, delete the action, submit
  await page.fill('input[name=description]', ruleNewDesc)
  await page.getByRole('button', { name: 'Next' }).click()
  await page
    .locator('div', { hasText: 'Create new alert' })
    .locator('[data-testid=CancelIcon]')
    .click()
  await expect(
    page.locator('#dialog-form').getByTestId('no-actions'),
  ).toBeVisible()
  await page.getByRole('button', { name: 'Next' }).click()
  await page.getByRole('button', { name: 'Submit' }).click()

  // see warning for no actions- check and submit
  await expect(page.getByText('WARNING: No actions')).toBeVisible()
  await page
    .locator('label', { hasText: 'I acknowledge the impact of this' })
    .locator('input[type=checkbox]')
    .click()
  await page.getByRole('button', { name: 'Retry' }).click()
  await expect(page.locator('[role=dialog]')).toBeHidden()

  // verify name does not change, with new description
  await expect(page.locator('body')).toContainText(ruleName)
  await expect(page.locator('body')).toContainText(ruleNewDesc)

  // delete the rule
  await page.getByRole('button', { name: 'Other Actions' }).click()
  await page.getByRole('menuitem', { name: 'Delete' }).click()
  expect(
    page.locator('[data-cy=dialog-title]', { hasText: 'Are you sure?' }),
  ).toBeVisible()
  await page.getByRole('button', { name: 'Confirm' }).click()

  // verify
  await expect(page.locator('[data-cy=list-empty-message]')).toHaveText(
    'No rules exist for this integration key.',
  )
})

test('create primary auth token then promote a second auth token', async ({
  page,
  playwright,
  isMobile,
}) => {
  const serviceURL = await setup(page, isMobile)

  await expect(page.locator('[data-cy=details]')).toContainText(
    'Auth Token: N/A',
  )

  await page.getByRole('button', { name: 'Generate Auth Token' }).click()
  await page.getByRole('button', { name: 'Generate' }).click()

  const origPrimaryToken = await page
    .getByRole('button', { name: 'Copy' })
    .textContent()
  expect(origPrimaryToken).not.toBeNull()
  await page.getByRole('button', { name: 'Done' }).click()

  function tokenHint(token: string): string {
    return token.slice(0, 2) + '...' + token.slice(-4)
  }

  await expect(page.locator('[data-cy=details]')).toContainText(
    tokenHint(origPrimaryToken as string),
  )

  const req = await playwright.request.newContext({
    baseURL: baseURLFromFlags(['univ-keys']),
    extraHTTPHeaders: { 'Content-Type': 'application/json', Cookie: '' },
  })

  async function testKey(key: string, status: number): Promise<void> {
    const auth = { Authorization: key ? 'Bearer ' + key : '' }
    const resp = await req.post('/api/v2/uik', { headers: auth, data: {} })
    expect(resp.status()).toBe(status)
  }

  await testKey('', 401)
  await testKey(origPrimaryToken as string, 204)

  await page.getByRole('button', { name: 'Generate Secondary Token' }).click()
  await page.getByRole('button', { name: 'Generate' }).click()
  const firstSecondaryToken = await page
    .getByRole('button', { name: 'Copy' })
    .textContent()
  expect(origPrimaryToken).not.toBeNull()
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
  expect(secondSecondaryToken).not.toBeNull()
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
