import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import PolicyStepFormDest, { FormValue } from './PolicyStepFormDest'
import { expect, userEvent, screen, waitFor, within } from '@storybook/test'
import { handleDefaultConfig, handleExpFlags } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { useArgs } from '@storybook/preview-api'

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
                vars.input.value === 'https://target.com2' ||
                vars.input.value === '+12225558989',
            },
          })
        }),

        graphql.query('DestDisplayInfo', ({ variables: vars }) => {
          let body = {}
          if (vars.input.values[0].value === 'https://target2.com') {
            body = {
              data: {
                destinationDisplayInfo: {
                  text: vars.input.values[0].value,
                  iconURL: 'builtin://webhook',
                  iconAltText: 'Webhook',
                },
              },
            }
          } else {
            body = {
              data: {
                destinationDisplayInfo: {
                  text: vars.input.values[0].value,
                  iconURL: 'builtin://phone-voice',
                  iconAltText: 'Voice Call',
                },
              },
            }
          }
          return HttpResponse.json(body)
        }),
      ],
    },
  },
} satisfies Meta<typeof PolicyStepFormDest>

export default meta
type Story = StoryObj<typeof meta>

export const AddActions: Story = {
  args: {
    value: {
      delayMinutes: 15,
      actions: [],
    },
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // add valid phone number and webhook
    await waitFor(async () => {
      await userEvent.type(
        await canvas.findByPlaceholderText('11235550123'),
        '12225558989',
      )
    })
    await expect(await canvas.findByTestId('CheckIcon')).toBeVisible()

    await userEvent.clear(
      await canvas.findByPlaceholderText('https://example.com'),
    )
    await waitFor(async () => {
      await userEvent.type(
        await canvas.findByPlaceholderText('https://example.com'),
        'https://target.com',
      )
    })

    await userEvent.click(await canvas.findByText('Add Action'))

    // expect to see action added
    await expect(await canvas.findByTestId('destination-chip')).toBeVisible()
    await expect(await canvas.findByText('+12225558989')).toBeVisible()

    // add single destination action
    await userEvent.click(await canvas.findByText('Multi Field EP Step Dest'))
    await userEvent.click(await screen.findByText('Single Field EP Step Dest'))
    await userEvent.clear(
      await canvas.findByPlaceholderText('https://example.com'),
    )
    await waitFor(async () => {
      await userEvent.type(
        await canvas.findByPlaceholderText('https://example.com'),
        'https://target2.com',
      )
    })

    await userEvent.click(await canvas.findByText('Add Action'))

    // expect to see action added
    await expect(await canvas.findByText('https://target2.com')).toBeVisible()
    await expect(await canvas.findByTestId('WebhookIcon')).toBeVisible()
  },
}
