import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import PolicyStepFormDest, { FormValue } from './PolicyStepFormDest'
import { expect, userEvent, waitFor, within, screen } from '@storybook/test'
import { handleDefaultConfig, handleExpFlags } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { useArgs } from '@storybook/preview-api'
import { DestFieldValueError, InputFieldError } from '../util/errtypes'

const meta = {
  title: 'escalation-policies/PolicyStepFormDest',
  component: PolicyStepFormDest,
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onChange = (newValue: FormValue): void => {
      if (args.onChange) args.onChange(newValue)
      setArgs({ value: newValue })
    }
    return <PolicyStepFormDest {...args} onChange={onChange} />
  },
  tags: ['autodocs'],
  parameters: {
    docs: {
      story: {
        inline: false,
        iframeHeight: 400,
      },
    },
    msw: {
      handlers: [
        handleDefaultConfig,
        handleExpFlags('dest-types'),
        graphql.query('ValidateDestination', ({ variables: vars }) => {
          return HttpResponse.json({
            data: {
              destinationFieldValidate:
                vars.input.value === 'https://target.com' ||
                vars.input.value === '+12225558989',
            },
          })
        }),

        graphql.query('DestDisplayInfo', ({ variables: vars }) => {
          console.log(vars)
          switch (vars.input.type) {
            case 'multi-field-ep-step':
              if (vars.input.values[0].value === '+123') {
                return HttpResponse.json({
                  data: null,
                  errors: [
                    {
                      message: 'number is too short',
                      path: ['destinationDisplayInfo', 'input'],
                      extensions: {
                        code: 'INVALID_DEST_FIELD_VALUE',
                        fieldID: 'phone-number',
                      },
                    } satisfies DestFieldValueError,
                    {
                      message: 'webhook url is invalid',
                      path: ['destinationDisplayInfo', 'input'],
                      extensions: {
                        code: 'INVALID_DEST_FIELD_VALUE',
                        fieldID: 'webhook-url',
                      },
                    } satisfies DestFieldValueError,
                  ],
                })
              }
              return HttpResponse.json({
                data: {
                  destinationDisplayInfo: {
                    text: vars.input.values[0].value,
                    iconURL: 'builtin://phone-voice',
                    iconAltText: 'Voice Call',
                  },
                },
              })
            case 'dest-type-error-ep-step':
              return HttpResponse.json({
                data: null,
                errors: [
                  {
                    message: 'This indicates an invalid destination type',
                    path: ['destinationDisplayInfo', 'input', 'type'],
                    extensions: {
                      code: 'INVALID_INPUT_VALUE',
                    },
                  } satisfies InputFieldError,
                ],
              })
            default:
              return HttpResponse.json({
                data: null,
                errors: [
                  {
                    message: 'This is a generic error',
                  },
                ],
              })
          }
        }),
      ],
    },
  },
} satisfies Meta<typeof PolicyStepFormDest>

export default meta
type Story = StoryObj<typeof meta>

export const AddAndDeleteAction: Story = {
  args: {
    value: {
      delayMinutes: 15,
      actions: [],
    },
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.click(await canvas.findByText('Dest Type Error EP Step'))
    await userEvent.click(await screen.findByText('Multi Field EP Step Dest'))

    // add action
    await userEvent.type(
      await canvas.findByPlaceholderText('11235550123'),
      '12225558989',
    )
    await userEvent.type(
      await canvas.findByPlaceholderText('https://example.com'),
      'https://target.com',
    )

    await userEvent.click(await canvas.findByText('Add Action'))

    // expect to see action added
    await expect(await canvas.findByText('+12225558989')).toBeVisible()
    await expect(await canvas.findByTestId('destination-chip')).toBeVisible()

    // delete action
    await userEvent.click(await canvas.findByTestId('CancelIcon'))

    // expect no actions
    await expect(await canvas.findByText('No actions')).toBeVisible()
  },
}

export const FieldErrors: Story = {
  args: {
    value: {
      delayMinutes: 15,
      actions: [],
    },
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.click(await canvas.findByText('Dest Type Error EP Step'))
    await userEvent.click(await screen.findByText('Multi Field EP Step Dest'))

    await userEvent.type(
      await canvas.findByPlaceholderText('11235550123'),
      '123',
    )
    await userEvent.type(
      await canvas.findByPlaceholderText('https://example.com'),
      'url',
    )

    await userEvent.click(await canvas.findByText('Add Action'))

    // expect to see fields are invalid
    await waitFor(async function InvalidField() {
      await expect(await canvas.findByLabelText('Phone Number')).toBeInvalid()
      await expect(await canvas.findByLabelText('Webhook URL')).toBeInvalid()
    })

    // add valid values
    await userEvent.clear(await canvas.findByPlaceholderText('11235550123'))
    await userEvent.type(
      await canvas.findByPlaceholderText('11235550123'),
      '12225558989',
    )
    await userEvent.clear(
      await canvas.findByPlaceholderText('https://example.com'),
    )
    await userEvent.type(
      await canvas.findByPlaceholderText('https://example.com'),
      'https://target.com',
    )

    // expect error messages to clear when editting text input
    await canvas.findByText(
      'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
    )
    await canvas.findByText('Webhook Documentation')
  },
}

export const TypeError: Story = {
  args: {
    value: {
      delayMinutes: 15,
      actions: [],
    },
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    await userEvent.type(
      await canvas.findByPlaceholderText('11235550123'),
      '456',
    )

    await userEvent.click(await canvas.findByText('Add Action'))

    // expect to see type error
    await expect(
      await canvas.findByText('This indicates an invalid destination type'),
    ).toBeVisible()
    await userEvent.click(await canvas.findByTestId('ErrorIcon'))
  },
}

export const GenericError: Story = {
  args: {
    value: {
      delayMinutes: 15,
      actions: [],
    },
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.click(await canvas.findByText('Dest Type Error EP Step'))
    await userEvent.click(await screen.findByText('Generic Error EP Step'))

    await userEvent.click(await canvas.findByText('Add Action'))

    // expect to see type error
    await expect(
      await canvas.findByText('This is a generic error'),
    ).toBeVisible()
    await userEvent.click(await canvas.findByTestId('ErrorIcon'))
  },
}
