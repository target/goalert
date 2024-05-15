import { test, expect } from '@playwright/test'
import { dropdownSelect, pageAction, userSessionFile } from './lib'
import Chance from 'chance'
const c = new Chance()

test.describe.configure({ mode: 'serial' })
test.use({ storageState: userSessionFile })

// test create, edit, verify, and delete of an EMAIL contact method
test('EMAIL contact method', async ({ page, browser }) => {
  const name = 'pw-email ' + c.name()
  const email = 'pw-email-' + c.email()

  await page.goto('./profile')

  await pageAction(page, 'Create Contact Method', 'Create Method')

  await page.fill('input[name=name]', name)
  await page.fill('input[name=type]', 'EMAIL')
  await page.fill('input[name=value]', email)
  await page.click('[role=dialog] button[type=submit]')

  const mail = await browser.newPage({
    baseURL: 'http://localhost:6125',
    viewport: { width: 800, height: 600 },
  })
  await mail.goto('./')
  await mail.fill('#search', email)
  await mail.press('#search', 'Enter')

  const message = mail.locator('.messages .msglist-message', {
    hasText: 'Verification Message',
  })
  await expect
    .poll(
      async () => {
        await mail.click('button[title=Refresh]')
        return await message.isVisible()
      },
      { message: 'wait for verification code email', timeout: 10000 },
    )
    .toBe(true)

  await message.click()

  const code = await mail
    .frameLocator('#preview-html')
    .locator('.invite-code')
    .textContent()
  if (!code) {
    throw new Error('No code found')
  }
  await mail.close()

  await page.fill('input[name=code]', code)
  await page.click('[role=dialog] button[type=submit]')
  await page.locator('[role=dialog]').isHidden()

  // edit name and enable status updates
  const updatedName = 'updated name ' + c.name()
  await page
    .locator('.MuiCard-root', {
      has: page.locator('div > div > h2', { hasText: 'Contact Methods' }),
    })
    .locator('li', { hasText: email })
    .locator('[aria-label="Other Actions"]')
    .click()
  await page.getByRole('menuitem', { name: 'Edit' }).click()
  await page.fill('input[name=name]', updatedName)
  await page.click('input[name=enableStatusUpdates]')
  await page.click('[role=dialog] button[type=submit]')

  // open edit dialog to verify name change and status updates are enabled
  await page
    .locator('.MuiCard-root', {
      has: page.locator('div > div > h2', { hasText: 'Contact Methods' }),
    })
    .locator('li', { hasText: email })
    .locator('[aria-label="Other Actions"]')
    .click()
  await page.getByRole('menuitem', { name: 'Edit' }).click()
  await expect(page.locator('input[name=name]')).toHaveValue(updatedName)
  await expect(page.locator('input[name=enableStatusUpdates]')).toBeChecked()
  await page.click('[role=dialog] button[type=submit]')

  // verify deleting a notification rule (immediate by default)
  await page
    .locator('li', {
      hasText: `Immediately notify me via Email at ${email}`,
    })
    .locator('button')
    .click()
  // click confirm
  await page.getByRole('button', { name: 'Confirm' }).click()
  await expect(
    page.locator('li', {
      hasText: `Immediately notify me via Email at ${email}`,
    }),
  ).not.toBeVisible()

  // verify adding a notification rule (delayed)
  await pageAction(page, 'Add Notification Rule', 'Add Rule')
  await dropdownSelect(page, 'Contact Method', updatedName)
  await page.fill('input[name=delayMinutes]', '5')
  await page.click('[role=dialog] button[type=submit]')

  await expect(
    page.locator('li', {
      hasText: `After 5 minutes notify me via Email at ${email}`,
    }),
  ).not.toBeVisible()

  await page
    .locator('.MuiCard-root', {
      has: page.locator('div > div > h2', { hasText: 'Contact Methods' }),
    })
    .locator('li', { hasText: email })
    .locator('[aria-label="Other Actions"]')
    .click()
  await page.getByRole('menuitem', { name: 'Delete' }).click()
  await page.getByRole('button', { name: 'Confirm' }).click()
  await page
    .locator('.MuiCard-root', {
      has: page.locator('div > div > h2', { hasText: 'Contact Methods' }),
    })
    .locator('li', { hasText: email })
    .isHidden()
})
