import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import PolicyStepFormDest, { FormValue } from './PolicyStepFormDest'
import { expect, userEvent, waitFor, within, fn } from '@storybook/test'
import { handleDefaultConfig } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { useArgs } from '@storybook/preview-api'
import { DestFieldValueError } from '../util/errtypes'

const VALID_PHONE = '+12225558989'
const VALID_PHONE2 = '+13335558989'
const INVALID_PHONE = '+15555'

const meta = {
  title: 'Escalation Policies/Steps/Form',
  component: PolicyStepFormDest,
  args: {
    onChange: fn(),
  },
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
    msw: {
      handlers: [
        handleDefaultConfig,
        graphql.query('ValidateDestination', ({ variables: vars }) => {
          return HttpResponse.json({
            data: {
              destinationFieldValidate: vars.input.value === VALID_PHONE,
            },
          })
        }),
        graphql.query('DestDisplayInfo', ({ variables: vars }) => {
          switch (vars.input.values[0].value) {
            case VALID_PHONE:
            case VALID_PHONE2:
              return HttpResponse.json({
                data: {
                  destinationDisplayInfo: {
                    text:
                      vars.input.values[0].value === VALID_PHONE
                        ? 'VALID_CHIP_1'
                        : 'VALID_CHIP_2',
                    iconURL: 'builtin://phone-voice',
                    iconAltText: 'Voice Call',
                  },
                },
              })
            default:
              return HttpResponse.json({
                errors: [
                  {
                    message: 'generic error',
                  },
                  {
                    path: ['destinationDisplayInfo', 'input'],
                    message: 'invalid phone number',
                    extensions: {
                      code: 'INVALID_DEST_FIELD_VALUE',
                      fieldID: 'phone-number',
                    },
                  } satisfies DestFieldValueError,
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
export const Empty: Story = {
  args: {
    value: {
      delayMinutes: 15,
      actions: [],
    },
  },
}

export const WithExistingActions: Story = {
  args: {
    value: {
      delayMinutes: 15,
      actions: [
        {
          type: 'single-field',
          values: [{ fieldID: 'phone-number', value: VALID_PHONE }],
        },
        {
          type: 'single-field',
          values: [{ fieldID: 'phone-number', value: VALID_PHONE2 }],
        },
      ],
    },
  },
}

export const ManageActions: Story = {
  args: {
    value: {
      delayMinutes: 15,
      actions: [],
    },
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const phoneInput = await canvas.findByLabelText('Phone Number')

    await userEvent.clear(phoneInput)
    await userEvent.type(phoneInput, INVALID_PHONE)
    await userEvent.click(await canvas.findByText('Add Destination'))

    await waitFor(async () => {
      await expect(await canvas.findByLabelText('Phone Number')).toBeInvalid()
      await expect(await canvas.findByText('generic error')).toBeVisible()
      await expect(
        await canvas.findByText('Invalid phone number'),
      ).toBeVisible()
    })

    await userEvent.clear(phoneInput)

    // Editing the input should clear the error
    await expect(await canvas.findByLabelText('Phone Number')).not.toBeInvalid()

    await userEvent.type(phoneInput, VALID_PHONE)

    await userEvent.click(await canvas.findByText('Add Destination'))

    // should result in chip
    await expect(await canvas.findByText('VALID_CHIP_1')).toBeVisible()

    // Delete the chip
    await userEvent.click(await canvas.findByTestId('CancelIcon'))

    await expect(await canvas.findByText('No actions')).toBeVisible()
  },
}
