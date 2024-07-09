import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodForm, { Value } from './UserContactMethodForm'
import { expect, within, userEvent, waitFor } from '@storybook/test'
import { handleDefaultConfig } from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'
import { HttpResponse, graphql } from 'msw'

const meta = {
  title: 'users/UserContactMethodForm',
  component: UserContactMethodForm,
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
    return <UserContactMethodForm {...args} onChange={onChange} />
  },
} satisfies Meta<typeof UserContactMethodForm>

export default meta
type Story = StoryObj<typeof meta>

export const SupportStatusUpdates: Story = {
  args: {
    value: {
      name: 'supports status',
      dest: {
        type: 'supports-status',
        args: { phone_number: '+15555555555' },
      },
      statusUpdates: false,
    },
    disabled: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    // ensure status updates checkbox is clickable

    await expect(
      await canvas.findByLabelText('Send alert status updates'),
    ).not.toBeDisabled()
  },
}

export const RequiredStatusUpdates: Story = {
  args: {
    value: {
      name: 'required status',
      dest: {
        type: 'required-status',
        args: { phone_number: '+15555555555' },
      },
      statusUpdates: false,
    },
    disabled: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    // ensure status updates checkbox is not clickable

    await expect(
      await canvas.findByLabelText(
        'Send alert status updates (cannot be disabled for this type)',
      ),
    ).toBeDisabled()
  },
}

export const ErrorSingleField: Story = {
  args: {
    value: {
      name: '-notvalid',
      dest: {
        type: 'single-field',
        args: { phone_number: '+15555555555' },
      },
      statusUpdates: false,
    },
    disabled: false,
    destFieldErrors: { phone_number: 'number is too short' },
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.type(await canvas.findByLabelText('Phone Number'), '123')

    // ensure errors are shown
    await expect(await canvas.findByText('Number is too short')).toBeVisible() // note: the 'N' is capitalized
    await waitFor(async function CloseIcon() {
      await expect(await canvas.findByTestId('CloseIcon')).toBeVisible()
    })
  },
}

export const ErrorMultiField: Story = {
  args: {
    value: {
      name: '-notvalid',
      dest: {
        type: 'triple-field',
        args: {
          'first-field': '+',
          'second-field': 'notAnEmail',
          'third-field': '-',
        },
      },
      statusUpdates: false,
    },
    disabled: false,
    nameError: 'must begin with a letter',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.type(await canvas.findByLabelText('First Item'), '123')

    // ensure errors are shown
    await expect(
      await canvas.findByText('Must begin with a letter'),
    ).toBeVisible()

    await waitFor(async function ThreeCloseIcons() {
      await expect(await canvas.findAllByTestId('CloseIcon')).toHaveLength(3)
    })
  },
}

export const Disabled: Story = {
  args: {
    value: {
      name: 'disabled dest',
      dest: {
        type: 'triple-field',
        args: {},
      },
      statusUpdates: false,
    },
    disabled: true,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    const combo = await canvas.findByRole('combobox')

    // get it's input field sibling (combo is a dom element)
    const input = combo.parentElement?.querySelector('input')
    await expect(input).toBeDisabled()

    await expect(
      await canvas.findByPlaceholderText('11235550123'),
    ).toBeDisabled()
    await expect(
      await canvas.findByPlaceholderText('foobar@example.com'),
    ).toBeDisabled()
    await expect(
      await canvas.findByPlaceholderText('slack user ID'),
    ).toBeDisabled()
    await expect(
      await canvas.findByLabelText('Send alert status updates'),
    ).toBeDisabled()
  },
}
