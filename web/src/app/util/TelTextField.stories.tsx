import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import TelTextField from './TelTextField'
import { expect, within } from '@storybook/test'
import { useArgs } from '@storybook/preview-api'

const meta = {
  title: 'util/TelTextField',
  component: TelTextField,
  argTypes: {
    component: { table: { disable: true } },
    ref: { table: { disable: true } },

    value: { control: 'text', defaultValue: '+17635550123' },
    label: { control: 'text', defaultValue: 'Phone Number' },
    error: { control: 'boolean' },
    onChange: { action: 'onChange' },
  },
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
      if (args.onChange) args.onChange(e)
      setArgs({ value: e.target.value })
    }
    return <TelTextField {...args} onChange={onChange} />
  },
  tags: ['autodocs'],
  parameters: {
    graphql: {
      PhoneNumberValidate: (vars: { number: string }) => ({
        data: {
          phoneNumberInfo: {
            id: vars.number,
            valid: vars.number.length === 12,
          },
        },
      }),
    },
  },
} satisfies Meta<typeof TelTextField>

export default meta
type Story = StoryObj<typeof meta>

export const ValidNumber: Story = {
  args: {
    value: '+17635550123',
    label: 'Phone Number',
    error: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    // ensure we get the red X

    await expect(await canvas.findByTestId('CheckIcon')).toBeVisible()
  },
}

export const InvalidNumber: Story = {
  args: {
    value: '+1763555012',
    label: 'Phone Number',
    error: false,
  },

  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    // ensure we get the red X

    await expect(await canvas.findByTestId('CloseIcon')).toBeVisible()
  },
}
