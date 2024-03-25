import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import PolicyStepEditDialogDest from './PolicyStepEditDialogDest'
import { expect, userEvent, screen, waitFor, fn } from '@storybook/test'
import { handleDefaultConfig, handleExpFlags } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { InputFieldError } from '../util/errtypes'
import { DestinationInput, UpdateEscalationPolicyStepInput } from '../../schema'

const meta = {
  title: 'escalation-policies/PolicyStepEditDialogDest',
  component: PolicyStepEditDialogDest,
  render: function Component(args) {
    return <PolicyStepEditDialogDest {...args} />
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
              destinationFieldValidate: vars.input.value === '+12225558989',
            },
          })
        }),
        graphql.query('DestDisplayInfo', ({ variables: vars }) => {
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
          'updateEscalationPolicyStep',
          ({ variables: vars }) => {
            console.log(vars)
            if (vars.input.escalationPolicyID === '1') {
              return HttpResponse.json({
                data: {
                  updateEscalationPolicyStep: true,
                },
              })
            }
            return HttpResponse.json({
              data: null,
              errors: [
                {
                  message: 'This is a generic input field error',
                  path: ['editEscalationPolicyStep', 'input', 'name'],
                  extensions: {
                    code: 'INVALID_INPUT_VALUE',
                  },
                } satisfies InputFieldError,
              ],
            })
          },
        ),
      ],
    },
  },
} satisfies Meta<typeof PolicyStepEditDialogDest>

const action: DestinationInput[] = [
  {
    type: 'builtin-twilio-sms',
    values: [{ fieldID: 'phone-number', value: '11235550123' }],
  },
]

const stepInput: UpdateEscalationPolicyStepInput = {
  actions: action,
  delayMinutes: 10,
  id: '1',
}

export default meta
type Story = StoryObj<typeof meta>

export const EditPolicyStep: Story = {
  argTypes: {
    onClose: { action: 'onClose' },
  },
  args: {
    escalationPolicyID: '1',
    step: stepInput,
    onClose: fn(),
  },
  play: async ({ args }) => {
    await userEvent.type(
      await screen.findByPlaceholderText('11235550123'),
      '12225558989',
    )

    await userEvent.click(await screen.findByText('Add Action'))

    await waitFor(async function Icon() {
      await userEvent.click(await screen.findByTestId('destination-chip'))
    })
    await userEvent.click(await screen.findByText('Submit'))

    await waitFor(async function Close() {
      expect(args.onClose).toHaveBeenCalled()
    })
  },
}

// export const GenericError: Story = {
//   args: {
//     escalationPolicyID: '2',
//     step: stepInput,
//     onClose: fn(),
//   },
//   play: async () => {
//     await userEvent.click(await screen.findByText('Dest Type Error EP Step'))
//     await userEvent.click(await screen.findByText('Generic Error EP Step'))
//     await userEvent.type(
//       await screen.findByPlaceholderText('11235550123'),
//       '12225558989',
//     )

//     await userEvent.click(await screen.findByText('Add Action'))
//     await userEvent.click(await screen.findByText('Submit'))

//     // expect error message
//     await expect(
//       await screen.findByText('This is a generic input field error'),
//     ).toBeVisible()
//   },
// }
