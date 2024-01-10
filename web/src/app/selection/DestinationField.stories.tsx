import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import DestinationField from './DestinationField'
import { expect } from '@storybook/jest'
import { within, userEvent } from '@storybook/testing-library'
import {
  handleDefaultConfig,
  handleDefaultDestTypes,
} from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'
import { HttpResponse, graphql } from 'msw'

const meta = {
  title: 'util/DestinationField',
  component: DestinationField,
  tags: ['autodocs'],
  argTypes: {
    destType: {
      control: 'select',
      options: [
        'builtin-webhook',
        'builtin-twilio-sms',
        'builtin-smtp-email',
        'builtin-slack-dm',
      ],
    },
  },
  parameters: {
    msw: {
      handlers: [
        handleDefaultConfig,
        handleDefaultDestTypes,
        graphql.query('ValidateDestination', ({ variables: vars }) => {
          return HttpResponse.json({
            data: {
              destinationFieldValidate:
                vars.input.value === 'https://test.com' ||
                vars.input.value === '+12225558989' ||
                vars.input.value === 'valid@email.com',
            },
          })
        }),
        graphql.query('DestinationSearchSelect', () => {
          return HttpResponse.json({
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
          })
        }),
        graphql.query('DestinationFieldValueName', ({ variables: vars }) => {
          return HttpResponse.json({
            data: {
              destinationFieldValueName:
                vars.input.value === 'C03SJES5FA7' ? '#general' : '',
            },
          })
        }),
      ],
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
} satisfies Meta<typeof DestinationInputDirect>

export default meta
type Story = StoryObj<typeof meta>

export const Webhook: Story = {
  args: {
    destType: 'builtin-webhook',
    value: [
      {
        fieldID: 'webhook-url',
        value: '',
      },
    ],
    disabled: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure placeholder and href loads correctly
    await expect(
      canvas.getByPlaceholderText('https://example.com'),
    ).toBeVisible()
    await expect(canvas.getByLabelText('Webhook URL')).toBeVisible()
    await expect(canvas.getByText('Webhook Documentation')).toHaveAttribute(
      'href',
      '/docs#webhooks',
    )

    // ensure check icon for valid URL
    await userEvent.clear(canvas.getByLabelText('Webhook URL'))
    await userEvent.type(
      canvas.getByLabelText('Webhook URL'),
      'https://test.com',
    )
    await expect(await canvas.findByTestId('CheckIcon')).toBeVisible()

    // ensure close icon for invalid URL
    await userEvent.clear(canvas.getByLabelText('Webhook URL'))
    await userEvent.type(
      canvas.getByLabelText('Webhook URL'),
      'not_a_valid_url',
    )
    await expect(await canvas.findByTestId('CloseIcon')).toBeVisible()
  },
}

export const PhoneNumbers: Story = {
  args: {
    destType: 'builtin-twilio-sms',
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

    // ensure check icon for valid number
    await userEvent.clear(canvas.getByLabelText('Phone Number'))
    await userEvent.type(canvas.getByLabelText('Phone Number'), '12225558989')
    await expect(await canvas.findByTestId('CheckIcon')).toBeVisible()

    // ensure close icon for invalid number
    await userEvent.clear(canvas.getByLabelText('Phone Number'))
    await userEvent.type(canvas.getByLabelText('Phone Number'), '123')
    await expect(await canvas.findByTestId('CloseIcon')).toBeVisible()

    // ensure only numbers are allowed
    await userEvent.clear(canvas.getByLabelText('Phone Number'))
    await userEvent.type(canvas.getByLabelText('Phone Number'), 'A4B5C6')
    await expect(
      canvas.getByLabelText('Phone Number').getAttribute('value'),
    ).toContain('456')
  },
}

export const Email: Story = {
  args: {
    destType: 'builtin-smtp-email',
    value: [
      {
        fieldID: 'email-address',
        value: '',
      },
    ],
    disabled: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure information renders correctly
    await expect(
      canvas.getByPlaceholderText('foobar@example.com'),
    ).toBeVisible()
    await expect(canvas.getByLabelText('Email Address')).toBeVisible()

    // ensure check icon for valid email
    await userEvent.clear(canvas.getByLabelText('Email Address'))
    await userEvent.type(
      canvas.getByLabelText('Email Address'),
      'valid@email.com',
    )
    await expect(await canvas.findByTestId('CheckIcon')).toBeVisible()

    // ensure close icon for invalid email
    await userEvent.clear(canvas.getByLabelText('Email Address'))
    await userEvent.type(canvas.getByLabelText('Email Address'), 'notvalid')
    await expect(await canvas.findByTestId('CloseIcon')).toBeVisible()
  },
}

export const SlackUserID: Story = {
  args: {
    destType: 'builtin-slack-dm',
    value: [
      {
        fieldID: 'slack-user-id',
        value: '',
      },
    ],
    disabled: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure information renders correctly
    await expect(canvas.getByPlaceholderText('member ID')).toBeVisible()
    await expect(
      canvas.getByText(
        'Go to your Slack profile, click the three dots, and select "Copy member ID".',
      ),
    ).toBeVisible()
  },
}

export const SlackChannel: Story = {
  args: {
    destType: 'builtin-slack-channel',
    value: [
      {
        fieldID: 'slack-channel-id',
        value: '',
      },
    ],
    disabled: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // should see #general channel as option
    await userEvent.type(
      canvas.getByPlaceholderText('Start typing...'),
      '#general',
    )
  },
}
