import { test } from '@playwright/test'

test.describe.configure({ mode: 'parallel' })

// test loging in with OIDC
test('OIDC Login', async ({ page }) => {
  await page.goto('./profile')

  await page.click('button[type=submit] >> "Login with OIDC"')

  // ensure we have an h1 with jane.doe
  await page.waitForSelector('h1 >> "jane.doe"')
})
