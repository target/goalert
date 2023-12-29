import React from 'react'
import { test, expect } from '@playwright/experimental-ct-react'
import { DestinationInput, Query } from '../../schema'
import DestinationInputChip from './DestinationInputChip'
import { DestInputChipValueWrapper } from './DestinationInputChip.story'
import { GQLMock } from '../../playwright/gqlMock'

test.use({ viewport: { width: 500, height: 500 } })

test('should render', async ({ mount, page }) => {
  const text = 'Corporate array Communications Rotation'

  const gql = new GQLMock(page)
  await gql.init()
  
  gql.setGQL('DestDisplayInfo', (vars) => ({
    data: {
        destinationDisplayInfo: {
            text,
            iconAltText: 'Rotation',
            iconURL: 'builtin://rotation',
            linkURL: 'test.com',
          },
    } as Query,
  }))

  const val: DestinationInput = {
    type: 'builtin-rotation',
    values: [{ fieldID: 'rotation-id', value: '123' }],
  }

  const component = await mount(
    <DestinationInputChip value={val} onDelete={() => {}} />,
  )
  await expect(component).toContainText(text)
  await expect(component).toHaveAttribute('href')
  await expect(
    component.locator('svg[data-testid="RotateRightIcon"]'),
  ).toBeVisible()
  await expect(component.locator('svg[data-testid="CancelIcon"]')).toBeVisible()
})

test('should delete', async ({ mount, page }) => {
  const text = 'Corporate array Communications Rotation'

  const gql = new GQLMock(page)
  await gql.init()
  
  gql.setGQL('DestDisplayInfo', (vars) => ({
    data: {
        destinationDisplayInfo: {
          text,
          iconAltText: 'Rotation',
          iconURL: 'builtin://rotation',
          linkURL: 'test.com',
        },
      } as Query,
  }))

  const component = await mount(<DestInputChipValueWrapper />)
  await expect(component).toContainText(text)
  await expect(
    component.locator('svg[data-testid="RotateRightIcon"]'),
  ).toBeVisible()
  component.locator('svg[data-testid="CancelIcon"]').click()
  await expect(
    component.locator('svg[data-testid="CancelIcon"]'),
  ).not.toBeVisible()
})
