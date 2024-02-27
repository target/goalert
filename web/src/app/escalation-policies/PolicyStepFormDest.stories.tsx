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
          console.log(vars)
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
            case 'multi-field-ep-step':
              if (vars.input.values[0].value === '+123') {
                return HttpResponse.json({
                  data: null,
                  errors: [
                    {
                      message: 'number is too short',
                      path: ['destinationDisplayInfo', 'input'],
                      extensions: {
                        code: 'INVALID_DEST_FIELD_VALUE',
                        fieldID: 'phone-number',
                      },
                    },
                    {
                      message: 'webhook url is invalid',
                      path: ['destinationDisplayInfo', 'input'],
                      extensions: {
                        code: 'INVALID_DEST_FIELD_VALUE',
                        fieldID: 'webhook-url',
                      },
                    },
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
      ],
    },
  },
} satisfies Meta<typeof PolicyStepFormDest>

export default meta
type Story = StoryObj<typeof meta>

export const AddActionAndDeleteActions: Story = {
  args: {
    value: {
      delayMinutes: 15,
      actions: [],
    },
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // add 'multi-field-ep-step' action
    await waitFor(async () => {
      await userEvent.type(
        await canvas.findByPlaceholderText('11235550123'),
        '12225558989',
      )
    })
    await expect(await canvas.findByTestId('CheckIcon')).toBeVisible()
    await waitFor(async () => {
      await userEvent.type(
        await canvas.findByPlaceholderText('https://example.com'),
        'https://target.com',
      )
    })

    await userEvent.click(await canvas.findByText('Add Action'))

    // expect to see action icon and label added
    await expect(await canvas.findByTestId('destination-chip')).toBeVisible()
    await expect(await canvas.findByText('+12225558989')).toBeVisible()

    // add single-field-ep-step action
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

    // expect to see action icon and label added
    await expect(await canvas.findByText('https://target2.com')).toBeVisible()
    await expect(await canvas.findByTestId('WebhookIcon')).toBeVisible()

    // delete all actions
    await userEvent.click(
      await canvas.findAllByTestId('CancelIcon').then((elem) => elem[0]),
    )
    await userEvent.click(await canvas.findByTestId('CancelIcon'))

    // expect no actions
    await expect(await canvas.findByText('No actions')).toBeVisible()
  },
}

export const Errors: Story = {
  args: {
    value: {
      delayMinutes: 15,
      actions: [],
    },
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    await waitFor(async () => {
      await userEvent.type(
        await canvas.findByPlaceholderText('11235550123'),
        '123',
      )
    })
    await waitFor(async () => {
      await userEvent.type(
        await canvas.findByPlaceholderText('https://example.com'),
        'notvalidurl',
      )
    })

    await userEvent.click(await canvas.findByText('Add Action'))

    // expect to see error messages
    await expect(await canvas.findByText('Number is too short')).toBeVisible()
    await expect(
      await canvas.findByText('Webhook url is invalid'),
    ).toBeVisible()

    // expect user input values to remain on textfield
    await expect(await canvas.findByDisplayValue('123')).toBeVisible()
    await expect(await canvas.findByDisplayValue('notvalidurl')).toBeVisible()

    // add valid values
    await userEvent.clear(await canvas.findByPlaceholderText('11235550123'))
    await waitFor(async () => {
      await userEvent.type(
        await canvas.findByPlaceholderText('11235550123'),
        '12225558989',
      )
    })
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

    // expect textfields to be empty

    await waitFor(async () => {
      await expect(
        await canvas.findByPlaceholderText('11235550123'),
      ).toHaveAttribute('value', '')
    })

    await waitFor(async () => {
      await expect(
        await canvas.findByPlaceholderText('https://example.com'),
      ).toHaveAttribute('value', '')
    })
  },
}
