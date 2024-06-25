import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import PolicyStepCreateDialog from './PolicyStepCreateDialog'
import { expect, fn, userEvent, waitFor, within } from '@storybook/test'
import { handleDefaultConfig } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { DestFieldValueError } from '../util/errtypes'

const meta = {
  title: 'Escalation Policies/Steps/Create Dialog',
  component: PolicyStepCreateDialog,
  render: function Component(args) {
    return <PolicyStepCreateDialog {...args} disablePortal />
  },
  tags: ['autodocs'],
  args: {
    onClose: fn(),
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

        graphql.mutation(
          'createEscalationPolicyStep',
          ({ variables: vars }) => {
            if (vars.input.delayMinutes === 999) {
              return HttpResponse.json({
                errors: [{ message: 'generic dialog error' }],
              })
            }

            return HttpResponse.json({
              data: {
                createEscalationPolicyStep: { id: '1' },
              },
            })
          },
        ),
      ],
    },
  },
} satisfies Meta<typeof PolicyStepCreateDialog>

export default meta
type Story = StoryObj<typeof meta>

export const CreatePolicyStep: Story = {
  argTypes: {
    onClose: { action: 'onClose' },
  },
  args: {
    escalationPolicyID: '1',
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
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

    await expect(await canvas.findByText('generic dialog error')).toBeVisible()

    await expect(args.onClose).not.toHaveBeenCalled() // should not close on error

    await userEvent.clear(delayField)
    await userEvent.type(delayField, '15')
    await userEvent.click(await canvas.findByText('Retry'))

    await waitFor(async function Close() {
      await expect(args.onClose).toHaveBeenCalled()
    })
  },
}
