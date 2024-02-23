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
                vars.input.value === 'https://target.com/webhook' ||
                vars.input.value === '+12225558989',
            },
          })
        }),
        graphql.query('DestDisplayInfo', ({ variables: vars }) => {
          let body = {}
          if (vars.input.type === 'single-field-ep-step') {
            body = {
              data: {
                destinationDisplayInfo: {
                  text: vars.input.values[0].value,
                  iconURL: 'builtin://webhook',
                  iconAltText: 'Webhook',
                  linkURL: '',
                },
              },
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
    // expect all fields to display correctly for single-field-ep-step
    await waitFor(async () => {
      await expect(await screen.findByLabelText('Webhook URL')).toBeVisible()
      await expect(
        await screen.findByPlaceholderText('https://example.com'),
      ).toBeVisible()
      await expect(screen.getByText('Webhook Documentation')).toHaveAttribute(
        'href',
        '/docs#webhooks',
      )
    })

    // add three actions
    await userEvent.clear(await screen.findByPlaceholderText('11235550123'))
    await waitFor(async () => {
      await userEvent.type(
        await screen.findByPlaceholderText('11235550123'),
        '12225558989',
      )
      await expect(await screen.findByTestId('CheckIcon')).toBeVisible()
    })

    // await userEvent.clear(
    //   await screen.findByPlaceholderText('https://example.com'),
    // )
    // await waitFor(async () => {
    //   await userEvent.type(
    //     await screen.findByPlaceholderText('https://example.com'),
    //     'https://target.com/webhook',
    //   )
    //   await expect(await screen.findByTestId('CheckIcon')).toBeVisible()
    // })
    // await userEvent.click(await screen.findByText('Add Action'))
    // await userEvent.clear(
    //   await screen.findByPlaceholderText('https://example.com'),
    // )
    // await waitFor(async () => {
    //   await userEvent.type(
    //     await screen.findByPlaceholderText('https://example.com'),
    //     'https://target.com/webhook2',
    //   )
    //   await expect(await screen.findByTestId('CheckIcon')).toBeVisible()
    // })
    // await userEvent.click(await screen.findByText('Add Action'))
  },
}
