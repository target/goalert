import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodFormDest, { Value } from './UserContactMethodFormDest'
import { expect } from '@storybook/jest'
import { within, screen, userEvent, waitFor } from '@storybook/testing-library'
import { handleDefaultConfig } from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'
import { HttpResponse, graphql } from 'msw'

const meta = {
  title: 'users/UserContactMethodFormDest',
  component: UserContactMethodFormDest,
  tags: ['autodocs'],
  parameters: {
    msw: {
      handlers: [
        handleDefaultConfig,
        graphql.query('ValidateDestination', ({ variables: vars }) => {
          return HttpResponse.json({
            data: {
              destinationFieldValidate: vars.input.value === '+15555555555',
            },
          })
        }),
      ],
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

export const SupportStatusUpdates: Story = {
  args: {
    value: {
      name: 'supports status',
      dest: {
        type: 'supports-status',
        values: [
          {
            fieldID: 'phone-number',
            value: '+15555555555',
          },
        ],
      },
      statusUpdates: false,
    },
    disabled: false,
  },
  play: async () => {
    // ensure status updates checkbox is clickable
    const status = await screen.getByLabelText('Send alert status updates')
    userEvent.click(status, {
      pointerEventsCheck: 1,
    })
  },
}

export const RequiredStatusUpdates: Story = {
  args: {
    value: {
      name: 'required status',
      dest: {
        type: 'required-status',
        values: [
          {
            fieldID: 'phone-number',
            value: '+15555555555',
          },
        ],
      },
      statusUpdates: false,
    },
    disabled: false,
  },
  play: async () => {
    // ensure status updates checkbox is not clickable
    const status = await screen.getByLabelText(
      'Send alert status updates (cannot be disabled for this type)',
    )
    userEvent.click(status, {
      pointerEventsCheck: 0,
    })
  },
}

export const ErrorSingleField: Story = {
  args: {
    value: {
      name: '-notvalid',
      dest: {
        type: 'single-field',
        values: [
          {
            fieldID: 'phone-number',
            value: '+',
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
    ],
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.type(screen.getByLabelText('Phone Number'), '123')

    // ensure errors are shown
    await expect(canvas.getByText('Must begin with a letter')).toBeVisible()
    await expect(await canvas.findByTestId('CloseIcon')).toBeVisible()
  },
}

export const ErrorMultiField: Story = {
  args: {
    value: {
      name: '-notvalid',
      dest: {
        type: 'triple-field',
        values: [
          {
            fieldID: 'first-field',
            value: '+',
          },
          {
            fieldID: 'second-field',
            value: 'notAnEmail',
          },
          {
            fieldID: 'third-field',
            value: '-',
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
    ],
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.type(screen.getByLabelText('First Item'), '123')

    // ensure errors are shown
    await expect(canvas.getByText('Must begin with a letter')).toBeVisible()
    await waitFor(async () => {
      await expect((await canvas.findAllByTestId('CloseIcon')).length).toBe(3)
    })
  },
}
