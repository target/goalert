import React from 'react'
import { test, expect } from '@playwright/experimental-ct-react'
import { TelTextValueWrapper } from './TelTextField.story'
import { GQLMock } from '../../playwright/gqlMock'

test.use({ viewport: { width: 500, height: 500 } })

test('should work', async ({ mount, page }) => {
  const gql = new GQLMock(page)
  await gql.init()

  gql.setGQL('TelTextInfo', (vars) => ({
    data: {
      phoneNumberInfo: {
        id: vars.number,
        valid: vars.number === '+17635550123',
      },
    },
  }))

  const component = await mount(<TelTextValueWrapper />)
  await component.locator('input').fill('17635550123')

  // ensure we have an SVG with attribute `data-testid="CheckIcon"`
  await expect(component.locator('svg[data-testid="CheckIcon"]')).toBeVisible()

  await component.locator('input').fill('123abc456')

  // value won't have the '+' because it's in the adornment, we should also
  // expect the input value to strip the non-numeric characters.
  await expect(component.locator('input')).toHaveValue('123456')

  await expect(component.locator('svg[data-testid="CloseIcon"]')).toBeVisible()
})
