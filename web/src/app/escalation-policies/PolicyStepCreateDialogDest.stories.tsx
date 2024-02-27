import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import PolicyStepCreateDialogDest from './PolicyStepCreateDialogDest'
import { expect, userEvent, screen, waitFor } from '@storybook/test'
import { handleDefaultConfig, handleExpFlags } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { InputFieldError } from '../util/errtypes'

const meta = {
  title: 'escalation-policies/PolicyStepCreateDialogDest',
  component: PolicyStepCreateDialogDest,
  render: function Component(args) {
    return <PolicyStepCreateDialogDest {...args} />
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
                vars.input.value === 'https://generic-error.com',
            },
          })
        }),
        graphql.query('DestDisplayInfo', ({ variables: vars }) => {
          switch (vars.input.type) {
            case 'single-field-ep-step':
              return HttpResponse.json({
                data: {
                  destinationDisplayInfo: {
                    text: vars.input.values[0].value,
                    iconURL: 'builtin://webhook',
                    iconAltText: 'Webhook',
                  },
                },
              })
            default:
              return HttpResponse.json({
                data: null,
                errors: [
                  {
                    message: 'destField is invalid',
                    path: ['destinationDisplayInfo', 'input', 'action'],
                    extensions: {
                      code: 'INVALID_DEST_FIELD_VALUE',
                      fieldID: 'phone-number',
                    },
                  },
                ],
              })
          }
        }),

        graphql.mutation(
          'createEscalationPolicyStep',
          ({ variables: vars }) => {
            console.log(vars)
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

export const GenericError: Story = {
  args: {
    escalationPolicyID: '00000000-0000-0000-0000-000000000000',
  },
  play: async () => {
    await userEvent.click(await screen.findByText('Multi Field EP Step Dest'))
    await userEvent.click(await screen.findByText('Single Field EP Step Dest'))

    await userEvent.clear(
      await screen.findByPlaceholderText('https://example.com'),
    )
    await waitFor(async () => {
      await userEvent.type(
        await screen.findByPlaceholderText('https://example.com'),
        'https://generic-error.com',
      )
    })

    await userEvent.click(await screen.findByText('Add Action'))
    await userEvent.click(await screen.findByText('Submit'))

    await expect(
      await screen.findByText('This is a generic input field error'),
    ).toBeVisible()
  },
}
