import { test, expect } from '@playwright/test'
import { baseURLFromFlags, userSessionFile } from './lib'

test.use({ storageState: userSessionFile })

test.describe(() => {
  // test a query for the current experimental flags (when example is set)
  test('example experimental flag set', async ({ page, isMobile }) => {
    test.skip(!!isMobile, 'mobile not supported for GraphQL explorer')

    await page.goto(baseURLFromFlags(['example']) + '/api/graphql/explore')

    await page.click('button[aria-label="Add tab"]')
    await page.click('.graphiql-editor')
    await page.keyboard.type(`{experimentalFlags`) // trailing curly brace will be added by the autocomplete
    await page.click('button.graphiql-execute-button')

    expect(page.locator('.result-window')).toContainText('experimentalFlags')

    const res = (await page.innerText('.result-window')).replace(/\s/g, '')
    expect(res).toContain(`{"data":{"experimentalFlags":["example"]}}`)
  })
})

// test a query for the current experimental flags (when none are set)
test('no experimental flags set', async ({ page, isMobile }) => {
  test.skip(!!isMobile, 'mobile not supported for GraphQL explorer')

  await page.goto('./api/graphql/explore')
  await page.click('.graphiql-editor')

  await page.click('button[aria-label="Add tab"]')
  await page.click('.graphiql-editor')
  await page.keyboard.type(`{experimentalFlags`) // trailing curly brace will be added by the autocomplete
  await page.click('button.graphiql-execute-button')

  expect(page.locator('.result-window')).toContainText('experimentalFlags')

  const res = (await page.innerText('.result-window')).replace(/\s/g, '')
  expect(res).toContain(`{"data":{"experimentalFlags":[]}}`)
})
