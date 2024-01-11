import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import DestinationField from './DestinationField'
import { expect } from '@storybook/jest'
import { within } from '@storybook/testing-library'
import { handleDefaultConfig } from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'
import { FieldValueInput } from '../../schema'

const meta = {
  title: 'util/DestinationField',
  component: DestinationField,
  tags: ['autodocs'],
  argTypes: {
    destType: {
      control: 'select',
      options: ['single-field', 'multi-field', 'disabled-destination'],
    },
  },
  parameters: {
    msw: {
      handlers: [handleDefaultConfig],
    },
  },
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onChange = (newValue: FieldValueInput[]): void => {
      if (args.onChange) args.onChange(newValue)
      setArgs({ value: newValue })
    }
    return <DestinationField {...args} onChange={onChange} />
  },
} satisfies Meta<typeof DestinationField>

export default meta
type Story = StoryObj<typeof meta>

export const SingleField: Story = {
  args: {
    destType: 'single-field',
    value: [
      {
        fieldID: 'phone-number',
        value: '',
      },
    ],
    disabled: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure information renders correctly
    await expect(canvas.getByLabelText('Phone Number')).toBeVisible()
    await expect(
      canvas.getByText(
        'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
      ),
    ).toBeVisible()
    await expect(canvas.getByText('+')).toBeVisible()
    await expect(canvas.getByPlaceholderText('11235550123')).toBeVisible()
  },
}

export const MultiField: Story = {
  args: {
    destType: 'multi-field',
    value: [
      {
        fieldID: 'phone-number',
        value: '',
      },
      {
        fieldID: 'email-address',
        value: '',
      },
    ],
    disabled: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure information for phone number renders correctly
    await expect(canvas.getByLabelText('Phone Number')).toBeVisible()
    await expect(
      canvas.getByText(
        'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
      ),
    ).toBeVisible()
    await expect(canvas.getByText('+')).toBeVisible()
    await expect(canvas.getByPlaceholderText('11235550123')).toBeVisible()

    // ensure information for email renders correctly
    await expect(
      canvas.getByPlaceholderText('foobar@example.com'),
    ).toBeVisible()
    await expect(canvas.getByLabelText('Email Address')).toBeVisible()

    // ensure information for slack renders correctly
    await expect(canvas.getByPlaceholderText('slack user ID')).toBeVisible()
    await expect(canvas.getByLabelText('Slack User')).toBeVisible()
  },
}

export const DisabledField: Story = {
  args: {
    destType: 'disabled-destination',
    value: [
      {
        fieldID: 'disabled',
        value: '',
      },
    ],
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure information renders correctly
    await expect(
      canvas.getByPlaceholderText('This field is disabled.'),
    ).toBeVisible()
  },
}
