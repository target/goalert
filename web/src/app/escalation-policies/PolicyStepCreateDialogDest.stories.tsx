import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import PolicyStepCreateDialogDest from './PolicyStepCreateDialogDest'
import { expect, userEvent, screen, waitFor, within } from '@storybook/test'
import { handleDefaultConfig, handleExpFlags } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { DestFieldValueError } from '../util/errtypes'

const meta = {
  title: 'Escalation Policies/Steps/Create Dialog',
  component: PolicyStepCreateDialogDest,
  render: function Component(args) {
    return <PolicyStepCreateDialogDest {...args} disablePortal />
  },
  tags: ['autodocs'],
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
        handleExpFlags('dest-types'),
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
} satisfies Meta<typeof PolicyStepCreateDialogDest>

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
    await userEvent.click(await screen.findByText('Add Action'))

    await expect(await canvas.findByText('Invalid number')).toBeVisible()
    await expect(await canvas.findByText('generic error')).toBeVisible()

    await userEvent.clear(phoneInput)
    await userEvent.type(phoneInput, '12225550123')
    await userEvent.click(await screen.findByText('Add Action'))

    await waitFor(async function Icon() {
      await userEvent.click(await screen.findByTestId('destination-chip'))
    })

    const delayField = await canvas.findByLabelText('Delay (minutes)')
    await userEvent.clear(delayField)
    await userEvent.type(delayField, '999')
    await userEvent.click(await screen.findByText('Submit'))

    await expect(await canvas.findByText('generic dialog error')).toBeVisible()

    await waitFor(async function Close() {
      expect(args.onClose).toHaveBeenCalled()
    })
  },
}

export const GenericError: Story = {
  args: {
    escalationPolicyID: '2',
  },
  play: async () => {
    await userEvent.click(await screen.findByText('Dest Type Error EP Step'))
    await userEvent.click(await screen.findByText('Generic Error EP Step'))
    await userEvent.type(
      await screen.findByPlaceholderText('11235550123'),
      '12225558989',
    )

    await userEvent.click(await screen.findByText('Add Action'))
    await userEvent.click(await screen.findByText('Submit'))

    // expect error message
    await expect(
      await screen.findByText('This is a generic input field error'),
    ).toBeVisible()
  },
}
