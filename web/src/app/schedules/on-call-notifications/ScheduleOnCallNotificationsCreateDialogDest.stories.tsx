import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { expect, fn, userEvent, waitFor, within } from '@storybook/test'
import ScheduleOnCallNotificationsCreateDialogDest from './ScheduleOnCallNotificationsCreateDialogDest'
import { mockOp } from '../../storybook/graphql'
import { BaseError, DestFieldValueError } from '../../util/errtypes'
import { SetScheduleOnCallNotificationRulesInput } from '../../../schema'

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
    fetchMock: {
      mocks: [
        mockOp<unknown, { id: string }>('SchedZone', (variables) => {
          return {
            data: {
              schedule: {
                id: variables.id,
                timeZone: 'America/Chicago',
              },
            },
          }
        }),
        mockOp('ValidateDestination', () => ({
          data: { destinationFieldValidate: true },
        })),
        mockOp<SetScheduleOnCallNotificationRulesInput>('SetRules', (vars) => {
          switch (vars.input.rules[0].dest?.values[0].value) {
            case '+123':
              return {
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
              }
            case '+1234567890':
              return {
                data: {
                  setScheduleOnCallNotificationRules: true,
                },
              }
          }

          throw new Error('unexpected value')
        }),
        mockOp<unknown, { scheduleID: string }>('GetRules', (variables) => {
          return {
            data: {
              schedule: {
                id: variables.scheduleID,
                onCallNotificationRules: [],
              },
            },
          }
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
