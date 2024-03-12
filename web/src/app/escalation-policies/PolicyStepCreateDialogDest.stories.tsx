import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import PolicyStepCreateDialogDest from './PolicyStepCreateDialogDest'
import { expect, userEvent, screen, waitFor } from '@storybook/test'
import { handleDefaultConfig, handleExpFlags } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { DestFieldValueError, InputFieldError } from '../util/errtypes'

const BAD_PHONE_NUMBER = '+12225558989'

const meta = {
  title: 'Escalation Policies/Steps/Create Dialog',
  component: PolicyStepCreateDialogDest,
  render: function Component(args) {
    return <PolicyStepCreateDialogDest {...args} />
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
          if (vars.input.value.length !== 12) {
            return HttpResponse.json({
              errors: [
                {
                  message: 'generic error',
                },
                {
                  message: 'Invalid number',
                  path: ['destinationDisplayInfo', 'input', 'dest'],
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
            console.log(vars)
            if (vars.input.escalationPolicyID === '1') {
              return HttpResponse.json({
                data: {
                  createEscalationPolicyStep: {
                    id: '1',
                    delayMinutes: 15,
                    targets: [
                      {
                        id: '11235550123',
                        name: '11235550123',
                        type: 'phone-number',
                        __typename: 'Target',
                      },
                    ],
                  },
                },
              })
            }
            return HttpResponse.json({
              data: null,
              errors: [
                {
                  message: 'This is a generic input field error',
                  path: ['createEscalationPolicyStep', 'input', 'name'],
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
  play: async ({ args }) => {
    await userEvent.click(await screen.findByText('Dest Type Error EP Step'))
    await userEvent.click(await screen.findByText('Multi Field EP Step Dest'))
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
