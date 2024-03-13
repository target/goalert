import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { expect, fn, userEvent, waitFor, within } from '@storybook/test'
import ScheduleOnCallNotificationsCreateDialogDest from './ScheduleOnCallNotificationsCreateDialogDest'
import { HttpResponse, graphql } from 'msw'
import { handleDefaultConfig, handleExpFlags } from '../../storybook/graphql'
import { BaseError, DestFieldValueError } from '../../util/errtypes'

const meta = {
  title: 'schedules/on-call-notifications/CreateDialogDest',
  component: ScheduleOnCallNotificationsCreateDialogDest,
  argTypes: {},
  args: {
    scheduleID: 'create-test',
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
        handleExpFlags('dest-types'),
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
                onCallNotificationRules: [],
              },
            },
          })
        }),
      ],
    },
  },
  tags: ['autodocs'],
  render: function Component(args) {
    return (
      <ScheduleOnCallNotificationsCreateDialogDest {...args} disablePortal />
    )
  },
} satisfies Meta<typeof ScheduleOnCallNotificationsCreateDialogDest>

export default meta
type Story = StoryObj<typeof meta>

export const Empty: Story = {
  args: {},
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
