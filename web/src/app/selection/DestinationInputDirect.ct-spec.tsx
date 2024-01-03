import React from 'react'
import { test, expect } from '@playwright/experimental-ct-react'
import { DestinationFieldConfig } from '../../schema'
import { DestinationInputDirectValueWrapper } from './DestinationInputDirect.story'
import { GQLMock } from '../../playwright/gqlMock'

test.use({ viewport: { width: 500, height: 500 } })

test('should render', async ({ mount, page }) => {
  const gql = new GQLMock(page)
  await gql.init()

  gql.setGQL('ValidateDestination', (vars) => {
    return {
      data: {
        destinationFieldValidate: vars.input.value === 'https://example.com',
      },
    }
  })

  const config: DestinationFieldConfig = {
    fieldID: 'webhook-url',
    hint: 'Webhook Documentation',
    hintURL: '/docs#webhooks',
    inputType: 'url',
    isSearchSelectable: false,
    labelPlural: 'Webhook URLs',
    labelSingular: 'Webhook URL',
    placeholderText: 'https://example.com',
    prefix: '',
    supportsValidation: true,
  }

  const component = await mount(
    <DestinationInputDirectValueWrapper
      value=' '
      config={config}
      destType='builtin-webhook'
    />,
  )
  // ensure text loads correctly
  await expect(component).toContainText('Webhook URL')
  await expect(component.locator('a')).toContainText('Webhook Documentation')
  await expect(component.locator('a')).toHaveAttribute('href')

  // ensure close icon visible for invalid urls
  await component.locator('input').fill('example')
  await expect(component.locator('input')).toHaveValue('example')
  await expect(component.locator('svg[data-testid="CloseIcon"]')).toBeVisible()

  // ensure check icon visible for valid urls
  await component.locator('input').fill('https://example.com')
  await expect(component.locator('input')).toHaveValue('https://example.com')
  await expect(component.locator('svg[data-testid="CheckIcon"]')).toBeVisible()
})
