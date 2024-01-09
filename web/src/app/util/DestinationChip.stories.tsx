import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import DestinationChip from './DestinationChip'
import { expect } from '@storybook/jest'
import { within } from '@storybook/testing-library'
import { handleDefaultConfig } from '../storybook/graphql'

const meta = {
  title: 'util/DestinationChip',
  component: DestinationChip,
  render: function Component(args) {
    return <DestinationChip {...args} />
  },
  tags: ['autodocs'],
  parameters: {
    msw: {
      handlers: [handleDefaultConfig],
    },
  },
} satisfies Meta<typeof DestinationChip>

export default meta
type Story = StoryObj<typeof meta>

export const TextAndHref: Story = {
  args: {
    config: {
      iconAltText: 'Schedule',
      iconURL: 'builtin://schedule',
      linkURL: 'test.com',
      text: 'Forward Integrated Functionality Schedule',
    },
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    await expect(
      canvas.getByText('Forward Integrated Functionality Schedule'),
    ).toBeVisible()

    await expect(canvas.getByTestId('destination-chip')).toHaveAttribute(
      'href',
      'test.com',
    )

    await expect(await canvas.findByTestId('TodayIcon')).toBeVisible()
    await expect(await canvas.findByTestId('CancelIcon')).toBeVisible()
  },
}

export const RotationIcon: Story = {
  args: {
    config: {
      iconAltText: 'Rotation',
      iconURL: 'builtin://rotation',
      linkURL: 'test.com',
      text: 'Icon Test',
    },
    onDelete: undefined,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(await canvas.findByTestId('RotateRightIcon')).toBeVisible()
  },
}

export const WebhookIcon: Story = {
  args: {
    config: {
      iconAltText: 'Webhook',
      iconURL: 'builtin://webhook',
      linkURL: 'test.com',
      text: 'Icon Test',
    },
    onDelete: undefined,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(await canvas.findByTestId('WebhookIcon')).toBeVisible()
  },
}

export const Error: Story = {
  args: {
    error: 'something went wrong',
    onDelete: undefined,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    await expect(canvas.getByText('ERROR: something went wrong')).toBeVisible()
    await expect(await canvas.findByTestId('BrokenImageIcon')).toBeVisible()
  },
}
