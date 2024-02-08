import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import DestinationInputDirect from './DestinationInputDirect'
import { expect, within, userEvent } from '@storybook/test'
import { handleDefaultConfig } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { useArgs } from '@storybook/preview-api'

const meta = {
  title: 'util/DestinationInputDirect',
  component: DestinationInputDirect,
  tags: ['autodocs'],
  argTypes: {
    inputType: {
      control: 'select',
      options: ['text', 'url', 'tel', 'email'],
      description: 'The type of input to use. tel will only allow numbers.',
    },
  },
  parameters: {
    msw: {
      handlers: [
        handleDefaultConfig,
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
      ],
    },
  },
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
      if (args.onChange) args.onChange(e)
      setArgs({ value: e.target.value })
    }
    return <DestinationInputDirect {...args} onChange={onChange} />
  },
} satisfies Meta<typeof DestinationInputDirect>

export default meta
type Story = StoryObj<typeof meta>

export const WebookWithDocLink: Story = {
  args: {
    value: '',

    fieldID: 'webhook-url',
    hint: 'Webhook Documentation',
    hintURL: '/docs#webhooks',
    inputType: 'url',
    labelSingular: 'Webhook URL',
    placeholderText: 'https://example.com',
    prefix: '',
    supportsValidation: true,
    isSearchSelectable: false,
    labelPlural: 'Webhook URLs',

    destType: 'builtin-webhook',
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
    value: '',

    fieldID: 'phone-number',
    hint: 'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
    hintURL: '',
    inputType: 'tel',
    labelSingular: 'Phone Number',
    placeholderText: '11235550123',
    prefix: '+',
    supportsValidation: true,
    labelPlural: 'Phone Numbers',
    isSearchSelectable: false,

    destType: 'builtin-twilio-sms',
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
    value: '',

    fieldID: 'email-address',
    hint: '',
    hintURL: '',
    inputType: 'email',
    labelSingular: 'Email Address',
    placeholderText: 'foobar@example.com',
    prefix: '',
    supportsValidation: true,
    isSearchSelectable: false,
    labelPlural: 'Email Addresses',

    destType: 'builtin-smtp-email',
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
