import { test, expect } from '@playwright/test'
import { userSessionFile } from './lib'
import Chance from 'chance'
const c = new Chance()

test.describe.configure({ mode: 'parallel' })
test.use({ storageState: userSessionFile })

// test create, edit, verify, and delete of an EMAIL contact method
test('EMAIL contact method', async ({ page, browser, isMobile }) => {
  const name = 'pw-email ' + c.name()
  const email = 'pw-email-' + c.email()

  await page.goto('./profile')

  if (isMobile) {
    await page.click('[aria-label="Add Items"]')
    await page.click('[aria-label="Create Contact Method"]')
  } else {
    await page.click('[title="Create Contact Method"]')
  }

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
  const updatedName = 'updated name'
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
