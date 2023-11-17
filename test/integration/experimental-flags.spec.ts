import { test, expect } from '@playwright/test'
import { baseURLFromFlags, userSessionFile } from './lib'

test.use({ storageState: userSessionFile })

// test a query for the current experimental flags (when example is set)
test('example experimental flag set', async ({ page }) => {
  await page.goto(baseURLFromFlags(['example']))
  await expect(page.locator('#content')).toHaveAttribute(
    'data-exp-flag-example',
    'true',
  )
})

// test a query for the current experimental flags (when none are set)
test('no experimental flags set', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('#content')).toHaveAttribute(
    'data-exp-flag-example',
    'false',
  )
})
