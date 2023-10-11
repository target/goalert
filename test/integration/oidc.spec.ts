import { test } from '@playwright/test'

test.describe.configure({ mode: 'parallel' })

// test logging in with OIDC
test('OIDC Login', async ({ page }) => {
  await page.goto('./profile')

  await page.click('button[type=submit] >> "Login with OIDC"')

  // ensure we have an h1 with jane.doe
  await page.waitForSelector('h1 >> "jane.doe"')
})
// test logging in with OIDC
test('OIDC Login (fallback public url)', async ({ page }) => {
  await page.goto('http://127.0.0.1:6120/profile')

  await page.click('button[type=submit] >> "Login with OIDC"')

  // ensure we have an h1 with jane.doe
  await page.waitForSelector('h1 >> "jane.doe"')
})
