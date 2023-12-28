import React from 'react'
import { test, expect } from '@playwright/experimental-ct-react'
import { TestWrapper } from './TelTextField.story'

test.use({ viewport: { width: 500, height: 500 } })

test('should work', async ({ mount, page }) => {
  await page.route('/api/graphql', (route) => {
    const body: { query: string; variables: { number: string } } = route
      .request()
      .postDataJSON()

    route.fulfill({
      status: 200,
      json: {
        data: {
          phoneNumberInfo: {
            id: body.variables.number,
            valid: body.variables.number === '+17635550123',
          },
        },
      },
    })
  })

  const component = await mount(<TestWrapper />)
  await component.locator('input').fill('17635550123')

  // ensure we have an SVG with attribute `data-testid="CheckIcon"`
  await expect(component.locator('svg[data-testid="CheckIcon"]')).toBeVisible()

  await component.locator('input').fill('123abc456')

  // value won't have the '+' because it's in the adornment, we should also
  // expect the input value to strip the non-numeric characters.
  await expect(component.locator('input')).toHaveValue('123456')

  await expect(component.locator('svg[data-testid="CloseIcon"]')).toBeVisible()
})
