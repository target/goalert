import { test, expect } from '@playwright/test'
import Chance from 'chance'
const c = new Chance()

test.describe.configure({ mode: 'parallel' })

// test loging in with OIDC
test('OIDC Login', async ({ page, browser, isMobile }) => {
  await page.goto('./profile')

  await page.click('button[type=submit] >> "Login with OIDC"')

  // ensure we have an h1 with jane.doe
  await page.waitForSelector('h1 >> "jane.doe"')
})
