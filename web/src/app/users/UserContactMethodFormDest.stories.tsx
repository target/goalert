import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodFormDest, { Value } from './UserContactMethodFormDest'
import { expect } from '@storybook/jest'
import { within } from '@storybook/testing-library'
import { handleDefaultConfig } from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'

const meta = {
  title: 'users/UserContactMethodFormDest',
  component: UserContactMethodFormDest,
  tags: ['autodocs'],
  parameters: {
    msw: {
      handlers: [handleDefaultConfig],
    },
  },
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onChange = (newValue: Value): void => {
      if (args.onChange) args.onChange(newValue)
      setArgs({ value: newValue })
    }
    return <UserContactMethodFormDest {...args} onChange={onChange} />
  },
} satisfies Meta<typeof UserContactMethodFormDest>

export default meta
type Story = StoryObj<typeof meta>

export const Error: Story = {
  args: {
    value: {
      name: '-notvalid',
      dest: {
        type: 'single-field',
        values: [
          {
            fieldID: 'phone-number',
            value: '+23',
          },
        ],
      },
      statusUpdates: false,
    },
    disabled: false,
    errors: [
      {
        field: 'name',
        message: 'must begin with a letter',
        name: 'FieldError',
        path: [],
        details: {},
      },
      {
        field: 'value',
        message:
          'must be a valid number: the phone number supplied is not a number',
        name: 'FieldError',
        path: [],
        details: {},
      },
    ],
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure error messages are shown
    await expect(canvas.getByText('Must begin with a letter')).toBeVisible()
    await expect(await canvas.findByTestId('CloseIcon')).toBeVisible()
  },
}
