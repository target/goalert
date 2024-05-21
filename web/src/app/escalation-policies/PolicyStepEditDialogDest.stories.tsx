import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import PolicyStepEditDialogDest from './PolicyStepEditDialogDest'
import { expect, fn, userEvent, waitFor, within } from '@storybook/test'
import { handleDefaultConfig } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { DestFieldValueError } from '../util/errtypes'
import { EscalationPolicyStep } from '../../schema'

const meta = {
  title: 'Escalation Policies/Steps/Edit Dialog',
  component: PolicyStepEditDialogDest,
  render: function Component(args) {
    return <PolicyStepEditDialogDest {...args} disablePortal />
  },
  tags: ['autodocs'],
  args: {
    onClose: fn(),
    escalationPolicyID: 'policy1',
    stepID: 'step1',
  },
  parameters: {
    docs: {
      story: {
        inline: false,
        iframeHeight: 600,
      },
    },
    msw: {
      handlers: [
        handleDefaultConfig,
        graphql.query('ValidateDestination', ({ variables: vars }) => {
          return HttpResponse.json({
            data: {
              destinationFieldValidate: vars.input.value.length === 12,
            },
          })
        }),
        graphql.query('DestDisplayInfo', ({ variables: vars }) => {
          if (vars.input.values[0].value.length !== 12) {
            return HttpResponse.json({
              errors: [
                { message: 'generic error' },
                {
                  message: 'Invalid number',
                  path: ['destinationDisplayInfo', 'input'],
                  extensions: {
                    code: 'INVALID_DEST_FIELD_VALUE',
                    fieldID: 'phone-number',
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
        }),

        graphql.query('GetEPStep', () => {
          return HttpResponse.json({
            data: {
              escalationPolicy: {
                id: 'policy1',
                steps: [
                  {
                    id: 'step1',
                    delayMinutes: 17,
                    actions: [
                      {
                        type: 'single-field',
                        values: [
                          { fieldID: 'phone-number', value: '+19995550123' },
                        ],
                      },
                    ],
                  } as EscalationPolicyStep,
                ],
              },
            },
          })
        }),

        graphql.mutation('UpdateEPStep', ({ variables: vars }) => {
          if (vars.input.delayMinutes === 999) {
            return HttpResponse.json({
              errors: [{ message: 'generic dialog error' }],
            })
          }

          return HttpResponse.json({
            data: {
              updateEscalationPolicyStep: true,
            },
          })
        }),
      ],
    },
  },
} satisfies Meta<typeof PolicyStepEditDialogDest>

export default meta
type Story = StoryObj<typeof meta>

export const UpdatePolicyStep: Story = {
  argTypes: {
    onClose: { action: 'onClose' },
  },
  args: {
    escalationPolicyID: '1',
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)

    // validate existing step data
    // 1. delay should be 17
    // 2. phone number should be +19995550123

    await waitFor(async function ExistingChip() {
      await expect(await canvas.findByLabelText('Delay (minutes)')).toHaveValue(
        17,
      )
      await expect(await canvas.findByText('+19995550123')).toBeVisible()
    })

    const phoneInput = await canvas.findByLabelText('Phone Number')
    await userEvent.clear(phoneInput)
    await userEvent.type(phoneInput, '1222')
    await userEvent.click(await canvas.findByText('Add Destination'))

    await expect(await canvas.findByText('Invalid number')).toBeVisible()
    await expect(await canvas.findByText('generic error')).toBeVisible()

    await userEvent.clear(phoneInput)
    await userEvent.type(phoneInput, '12225550123')
    await userEvent.click(await canvas.findByText('Add Destination'))

    await waitFor(async function Icon() {
      await userEvent.click(await canvas.findByTestId('destination-chip'))
    })

    const delayField = await canvas.findByLabelText('Delay (minutes)')
    await userEvent.clear(delayField)
    await userEvent.type(delayField, '999')
    await userEvent.click(await canvas.findByText('Submit'))

    await waitFor(async function Error() {
      await expect(
        await canvas.findByText('generic dialog error'),
      ).toBeVisible()
    })

    await expect(args.onClose).not.toHaveBeenCalled() // should not close on error

    await userEvent.clear(delayField)
    await userEvent.type(delayField, '15')
    await userEvent.click(await canvas.findByText('Retry'))

    await waitFor(async function Close() {
      await expect(args.onClose).toHaveBeenCalled()
    })
  },
}
