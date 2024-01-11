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
  argTypes: {
    iconURL: {
      control: 'select',
      options: [
        'builtin://schedule',
        'builtin://rotation',
        'builtin://webhook',
        'builtin://slack',
      ],
    },
    onDelete: {
      control: 'select',
      options: [() => null, undefined],
    },
  },
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
    iconAltText: 'Schedule',
    iconURL: 'builtin://schedule',
    linkURL: 'https://example.com',
    text: 'Forward Integrated Functionality Schedule',

    onDelete: () => null,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    await expect(
      canvas.getByText('Forward Integrated Functionality Schedule'),
    ).toBeVisible()

    await expect(canvas.getByTestId('destination-chip')).toHaveAttribute(
      'href',
      'https://example.com',
    )

    await expect(await canvas.findByTestId('TodayIcon')).toBeVisible()
    await expect(await canvas.findByTestId('CancelIcon')).toBeVisible()
  },
}

export const RotationIcon: Story = {
  args: {
    iconAltText: 'Rotation',
    iconURL: 'builtin://rotation',
    linkURL: 'https://example.com',
    text: 'Icon Test',

    onDelete: undefined,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(await canvas.findByTestId('RotateRightIcon')).toBeVisible()
  },
}

export const WebhookIcon: Story = {
  args: {
    iconAltText: 'Webhook',
    iconURL: 'builtin://webhook',
    linkURL: 'https://example.com',
    text: 'Icon Test',

    onDelete: undefined,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(await canvas.findByTestId('WebhookIcon')).toBeVisible()
  },
}

export const Loading: Story = {
  args: {
    iconAltText: 'Webhook',
    iconURL: 'builtin://webhook',
    linkURL: 'https://example.com',
    text: '',

    onDelete: undefined,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(await canvas.findByTestId('spinner')).toBeVisible()
  },
}

export const NoIcon: Story = {
  args: {
    iconAltText: '',
    iconURL: '',
    linkURL: '',
    text: 'No Icon Test',

    onDelete: undefined,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(canvas.getByText('No Icon Test')).toBeVisible()
  },
}

export const Error: Story = {
  args: {
    iconAltText: '',
    iconURL: '',
    linkURL: '',
    text: '',
    error: 'something went wrong',
    onDelete: undefined,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    await expect(canvas.getByText('ERROR: something went wrong')).toBeVisible()
    await expect(await canvas.findByTestId('BrokenImageIcon')).toBeVisible()
  },
}
