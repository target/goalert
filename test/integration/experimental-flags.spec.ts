import { test, expect } from '@playwright/test'
import { configureExpFlags, userSessionFile } from './lib'

test.use({ storageState: userSessionFile })

test.describe(() => {
  configureExpFlags(['example'])

  // test a query for the current experimental flags (when example is set)
  test('example experimental flag set', async ({ page, browser, isMobile }) => {
    await page.goto('./api/graphql/explore')
    await page.click('.graphiql-editor')
    await page.keyboard.down('Control')
    await page.keyboard.press('KeyA')
    await page.keyboard.up('Control')
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
test('no experimental flags set', async ({ page, browser, isMobile }) => {
  await page.goto('./api/graphql/explore')
  await page.click('.graphiql-editor')
  await page.keyboard.down('Control')
  await page.keyboard.press('KeyA')
  await page.keyboard.up('Control')
  await page.keyboard.type(`{experimentalFlags`) // trailing curly brace will be added by the autocomplete

  await page.keyboard.down('Control')
  await page.keyboard.press('Enter')
  await page.keyboard.up('Control')

  expect(page.locator('.result-window')).toContainText('experimentalFlags')

  const res = (await page.innerText('.result-window')).replace(/\s/g, '')
  expect(res).toContain(`{"data":{"experimentalFlags":[]}}`)
})
