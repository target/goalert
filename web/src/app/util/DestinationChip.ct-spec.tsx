import React from 'react'
import { test, expect } from '@playwright/experimental-ct-react'
import { DestinationDisplayInfo } from '../../schema'
import DestinationChip from './DestinationChip'

test.use({ viewport: { width: 500, height: 500 } })

test('should render', async ({ mount }) => {
  const text = 'Forward Integrated Functionality Schedule'
  const config: DestinationDisplayInfo = {
    iconAltText: 'Schedule',
    iconURL: 'builtin://schedule',
    linkURL: 'test.com',
    text,
  }

  const component = await mount(
    <DestinationChip config={config} onDelete={() => {}} />,
  )
  await expect(component).toContainText(text)
  await expect(component).toHaveAttribute('href')
  await expect(component.locator('svg[data-testid="TodayIcon"]')).toBeVisible()
  await expect(component.locator('svg[data-testid="CancelIcon"]')).toBeVisible()
})

test('should display error', async ({ mount }) => {
  const component = await mount(
    <DestinationChip error='something went wrong' />,
  )
  await expect(component).toContainText('ERROR: something went wrong')
  await expect(
    component.locator('svg[data-testid="BrokenImageIcon"]'),
  ).toBeVisible()
})
