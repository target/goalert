import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import PolicyStepCreateDialogDest from './PolicyStepCreateDialogDest'
import { expect, userEvent, screen, waitFor } from '@storybook/test'
import { handleDefaultConfig, handleExpFlags } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'

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
                vars.input.value === 'https://target.com' ||
                vars.input.value === 'https://target.com2' ||
                vars.input.value === '+12225558989',
            },
          })
        }),

        graphql.query('DestDisplayInfo', ({ variables: vars }) => {
          let body = {}
          if (vars.input.values[0].value === '+111') {
            body = {
              data: null,
              errors: [
                {
                  message: 'phone number is invalid',
                  path: ['destinationDisplayInfo', 'input', 'action'],
                  extensions: {
                    code: 'INVALID_DEST_FIELD_VALUE',
                    fieldID: 'phone-number',
                  },
                },
              ],
            }
          } else if (
            vars.input.values[0].value === '+12225558989' &&
            vars.input.values[1].value === 'notvalid'
          ) {
            body = {
              data: null,
              errors: [
                {
                  message: 'webhook is invalid',
                  path: ['destinationDisplayInfo', 'input', 'action'],
                  extensions: {
                    code: 'INVALID_DEST_FIELD_VALUE',
                    fieldID: 'webhook-url',
                  },
                },
              ],
            }
          } else if (vars.input.values[0].value === 'https://target2.com') {
            body = {
              data: {
                destinationDisplayInfo: {
                  text: vars.input.values[0].value,
                  iconURL: 'builtin://webhook',
                  iconAltText: 'Webhook',
                },
              },
            }
          } else if (vars.input.values[0].value === '+12225558989') {
            body = {
              data: {
                destinationDisplayInfo: {
                  text: vars.input.values[0].value,
                  iconURL: 'builtin://phone-voice',
                  iconAltText: 'Voice Call',
                },
              },
            }
          } else {
            body = {
              errors: [
                {
                  message: 'invalid dest input',
                  path: ['destinationDisplayInfo', 'input', 'action'],
                  extensions: {
                    code: 'INVALID_DEST_FIELD_VALUE',
                    fieldID: vars.input.values[0].type || 'generic',
                  },
                },
              ],
            }
          }
          return HttpResponse.json(body)
        }),
      ],
    },
  },
} satisfies Meta<typeof PolicyStepCreateDialogDest>

export default meta
type Story = StoryObj<typeof meta>

export const AddActions: Story = {
  args: {
    escalationPolicyID: '00000000-0000-0000-0000-000000000000',
  },
  play: async () => {
    // add invalid phone number
    await userEvent.clear(await screen.findByPlaceholderText('11235550123'))
    await waitFor(async () => {
      await userEvent.type(
        await screen.findByPlaceholderText('11235550123'),
        '111',
      )
    })
    await userEvent.click(await screen.findByText('Add Action'))

    // expect to see error message for phone number
    await expect(
      await screen.findByText('phone number is invalid'),
    ).toBeVisible()

    await userEvent.clear(await screen.findByPlaceholderText('11235550123'))
    await waitFor(async () => {
      await userEvent.type(
        await screen.findByPlaceholderText('11235550123'),
        '12225558989',
      )
    })
    await expect(await screen.findByTestId('CheckIcon')).toBeVisible()

    // add invalid webhook
    await waitFor(async () => {
      await userEvent.type(
        await screen.findByPlaceholderText('https://example.com'),
        'notvalid',
      )
    })

    await userEvent.click(await screen.findByText('Add Action'))

    // expect to see error message for webhook
    await expect(await screen.findByText('webhook is invalid')).toBeVisible()

    // add valid phone number and webhook
    await waitFor(async () => {
      await userEvent.type(
        await screen.findByPlaceholderText('11235550123'),
        '12225558989',
      )
    })
    await expect(await screen.findByTestId('CheckIcon')).toBeVisible()

    await userEvent.clear(
      await screen.findByPlaceholderText('https://example.com'),
    )
    await waitFor(async () => {
      await userEvent.type(
        await screen.findByPlaceholderText('https://example.com'),
        'https://target.com',
      )
    })

    await userEvent.click(await screen.findByText('Add Action'))

    // expect to see action added
    await expect(await screen.findByTestId('destination-chip')).toBeVisible()
    await expect(await screen.findByText('+12225558989')).toBeVisible()

    // add single destination action
    await userEvent.click(await screen.findByText('Multi Field EP Step Dest'))
    await userEvent.click(await screen.findByText('Single Field EP Step Dest'))
    await userEvent.clear(
      await screen.findByPlaceholderText('https://example.com'),
    )
    await waitFor(async () => {
      await userEvent.type(
        await screen.findByPlaceholderText('https://example.com'),
        'https://target2.com',
      )
    })

    await userEvent.click(await screen.findByText('Add Action'))

    // expect to see action added
    await expect(await screen.findByText('https://target2.com')).toBeVisible()
    await expect(await screen.findByTestId('WebhookIcon')).toBeVisible()
  },
}
