import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { expect, fn, userEvent, waitFor, within } from '@storybook/test'
import ScheduleOnCallNotificationsEditDialogDest from './ScheduleOnCallNotificationsEditDialogDest'
import { HttpResponse, graphql } from 'msw'
import { handleDefaultConfig } from '../../storybook/graphql'
import { BaseError, DestFieldValueError } from '../../util/errtypes'

const meta = {
  title: 'schedules/on-call-notifications/EditDialogDest',
  component: ScheduleOnCallNotificationsEditDialogDest,
  argTypes: {},
  args: {
    scheduleID: 'create-test',
    ruleID: 'existing-id',
    onClose: fn(),
  },
  parameters: {
    docs: {
      story: {
        inline: false,
        iframeHeight: 500,
      },
    },
    msw: {
      handlers: [
        handleDefaultConfig,
        graphql.query('SchedZone', ({ variables }) => {
          return HttpResponse.json({
            data: {
              schedule: {
                id: variables.id,
                timeZone: 'America/Chicago',
              },
            },
          })
        }),
        graphql.query('ValidateDestination', () =>
          HttpResponse.json({ data: { destinationFieldValidate: true } }),
        ),
        graphql.mutation('SetRules', ({ variables }) => {
          switch (variables.input.rules[0].dest.values[0].value) {
            case '+123':
              return HttpResponse.json({
                errors: [
                  {
                    message: 'Generic Error',
                  } satisfies BaseError,
                  {
                    message: 'field error',
                    path: [
                      'setScheduleOnCallNotificationRules',
                      'input',
                      'rules',
                      0,
                      'dest',
                    ],
                    extensions: {
                      code: 'INVALID_DEST_FIELD_VALUE',
                      fieldID: 'phone-number',
                    },
                  } satisfies DestFieldValueError,
                ],
              })
            case '+1234567890':
              return HttpResponse.json({
                data: {
                  setScheduleOnCallNotificationRules: true,
                },
              })
          }

          throw new Error('unexpected value')
        }),
        graphql.query('GetRules', ({ variables }) => {
          return HttpResponse.json({
            data: {
              schedule: {
                id: variables.scheduleID,
                onCallNotificationRules: [
                  {
                    id: 'existing-id',
                    dest: {
                      type: 'single-field',
                      values: [
                        {
                          fieldID: 'phone-number',
                          value: '+1234567890',
                        },
                      ],
                    },
                  },
                ],
              },
            },
          })
        }),
      ],
    },
  },
  tags: ['autodocs'],
  render: function Component(args) {
    return <ScheduleOnCallNotificationsEditDialogDest {...args} disablePortal />
  },
} satisfies Meta<typeof ScheduleOnCallNotificationsEditDialogDest>

export default meta
type Story = StoryObj<typeof meta>

export const ExistingRule: Story = {
  args: {},
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(await canvas.findByLabelText('Phone Number')).toHaveValue(
      '1234567890', // input field won't include the "+" prefix
    )
  },
}

export const ValidationError: Story = {
  args: {},
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    await userEvent.clear(await canvas.findByLabelText('Phone Number'))
    await userEvent.type(await canvas.findByLabelText('Phone Number'), '123')

    await userEvent.click(await canvas.findByText('Submit'))

    await waitFor(
      async () => {
        await expect(await canvas.findByLabelText('Phone Number')).toBeInvalid()
        await expect(await canvas.findByText('Field error')).toBeVisible()
        await expect(await canvas.findByText('Generic Error')).toBeVisible()
      },
      { timeout: 5000 },
    )

    await userEvent.clear(await canvas.findByLabelText('Phone Number'))
    await userEvent.type(
      await canvas.findByLabelText('Phone Number'),
      '1234567890',
    )

    await waitFor(async () => {
      // editing the input field should clear errors
      await expect(
        await canvas.findByLabelText('Phone Number'),
      ).not.toBeInvalid()
    })
  },
}
