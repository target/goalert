import React from 'react'
import { test, expect } from '@playwright/experimental-ct-react'
import { DestinationFieldConfig } from '../../schema'
import { GQLMock } from '../../playwright/gqlMock'
import { DestinationSearchSelectWrapper } from './DestinationSearchSelect.story'

test.use({ viewport: { width: 500, height: 500 } })

test('should render', async ({ mount, page }) => {
  const gql = new GQLMock(page)
  await gql.init()

  gql.setGQL('DestinationSearchSelect', (vars) => {
    return {
      data: {
        destinationFieldSearch: {
          nodes: [
            {
              value: 'C03SJES5FA7',
              label: '#general',
              isFavorite: false,
              __typename: 'FieldValuePair',
            },
          ],
          __typename: 'FieldValueConnection',
        },
      },
    }
  })

  gql.setGQL('DestinationFieldValueName', (vars) => {
    return {
      data: {
        destinationFieldValueName:
          vars.input.value === 'C03SJES5FA7' ? '#general' : '',
      },
    }
  })

  const config: DestinationFieldConfig = {
    fieldID: 'slack-channel-id',
    hint: '',
    hintURL: '',
    inputType: 'text',
    isSearchSelectable: true,
    labelPlural: 'Slack Channels',
    labelSingular: 'Slack Channel',
    placeholderText: '',
    prefix: '',
    supportsValidation: false,
  }

  const component = await mount(
    <DestinationSearchSelectWrapper
      value=' '
      config={config}
      destType='builtin-slack-channel'
    />,
  )
  // ensure text loads correctly
  await expect(component).toContainText('Slack Channel')
  await component.locator('input').fill('#gen')

  // await expect(page.locator('ul')).toHaveText('#general')

  // await expect(component.locator('ul')).toContainText('#general')

  // await expect(component).toContainText('#general')

  await expect(component.locator('span[data-option-index=0]')).toContainText(
    '#general',
  )

  // input[name=username]
})
