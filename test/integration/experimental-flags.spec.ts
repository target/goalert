import { test, expect } from '@playwright/test'
import { baseURLFromFlags, userSessionFile } from './lib'

import { platform } from 'os'
const isMacOS = platform() === 'darwin'

test.use({ storageState: userSessionFile })

test.describe(() => {
  // test a query for the current experimental flags (when example is set)
  test('example experimental flag set', async ({ page }) => {
    await page.goto(baseURLFromFlags(['example']) + '/api/graphql/explore')
    await page.click('.graphiql-editor')

    await page.keyboard.press(isMacOS ? 'Meta+A' : 'Control+A')
    await page.keyboard.type(`{experimentalFlags`) // trailing curly brace will be added by the autocomplete

    await page.keyboard.down('Control')
    await page.keyboard.press('Enter')
    await page.keyboard.up('Control')

    expect(page.locator('.result-window')).toContainText('experimentalFlags')

    const res = (await page.innerText('.result-window')).replace(/\s/g, '')
    expect(res).toContain(`{"data":{"experimentalFlags":["example"]}}`)
  })
})

// test a query for the current experimental flags (when none are set)
test('no experimental flags set', async ({ page }) => {
  await page.goto('./api/graphql/explore')
  await page.click('.graphiql-editor')

  await page.keyboard.press(isMacOS ? 'Meta+A' : 'Control+A')
  await page.keyboard.type(`{experimentalFlags`) // trailing curly brace will be added by the autocomplete

  await page.keyboard.down('Control')
  await page.keyboard.press('Enter')
  await page.keyboard.up('Control')

  expect(page.locator('.result-window')).toContainText('experimentalFlags')

  const res = (await page.innerText('.result-window')).replace(/\s/g, '')
  expect(res).toContain(`{"data":{"experimentalFlags":[]}}`)
})
