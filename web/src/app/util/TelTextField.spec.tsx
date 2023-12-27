import React from 'react'
import { test, expect } from '@playwright/experimental-ct-react'
import TelTextField from './TelTextField'

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

  let newValue = ''

  let component = await mount(
    <TelTextField
      value='17635550123'
      onChange={(v) => {
        newValue = v.target.value
      }}
    />,
  )
  // ensure we have an SVG with attribute `data-testid="CheckIcon"`

  await expect(component.locator('svg[data-testid="CheckIcon"]')).toBeVisible()

  await component.locator('input').fill('123abc456')

  expect(newValue).toBe('+123456') // should be stripped of non-numeric characters

  await component.unmount()

  component = await mount(<TelTextField value='1111' />)
  await expect(component.locator('svg[data-testid="CloseIcon"]')).toBeVisible()
})
