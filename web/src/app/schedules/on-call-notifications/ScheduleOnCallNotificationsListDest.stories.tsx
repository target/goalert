import type { Meta, StoryObj } from '@storybook/react'
import ScheduleOnCallNotificationsListDest from './ScheduleOnCallNotificationsListDest'
import { mockOp } from '../../storybook/graphql'

const emptyScheduleID = '00000000-0000-0000-0000-000000000000'
const errorScheduleID = '11111111-1111-1111-1111-111111111111'
const manyNotificationsScheduleID = '22222222-2222-2222-2222-222222222222'
type SchedVar = { scheduleID: string }
const meta = {
  title: 'schedules/on-call-notifications/ListDest',
  component: ScheduleOnCallNotificationsListDest,
  argTypes: {},
  parameters: {
    fetchMock: {
      mocks: [
        mockOp<unknown, SchedVar>('ScheduleNotifications', (vars) => {
          switch (vars.scheduleID) {
            case emptyScheduleID:
              return {
                data: {
                  schedule: {
                    id: emptyScheduleID,
                    timeZone: 'America/Chicago',
                    onCallNotificationRules: [],
                  },
                },
              }
            case errorScheduleID:
              return {
                errors: [
                  { message: 'This is an example of a returned GraphQL error' },
                ],
              }
            case manyNotificationsScheduleID:
              return {
                data: {
                  schedule: {
                    id: manyNotificationsScheduleID,
                    timeZone: 'America/Chicago',
                    onCallNotificationRules: [
                      {
                        id: '1',
                        time: '08:00',
                        weekdayFilter: [
                          true,
                          true,
                          true,
                          true,
                          true,
                          true,
                          true,
                        ],
                        dest: {
                          displayInfo: {
                            text: 'example.com',
                            iconURL: 'builtin://webhook',
                            iconAltText: 'Webhook',
                          },
                        },
                      },
                      {
                        id: '2',
                        time: '08:00',
                        weekdayFilter: [
                          false,
                          false,
                          false,
                          false,
                          false,
                          false,
                          false,
                        ],
                        dest: {
                          displayInfo: {
                            text: 'other.example.com',
                            iconURL: 'builtin://webhook',
                            iconAltText: 'Webhook',
                          },
                        },
                      },
                    ],
                  },
                },
              }

            default:
              throw new Error('Unknown scheduleID')
          }
        }),
      ],
    },
  },
  tags: ['autodocs'],
} satisfies Meta<typeof ScheduleOnCallNotificationsListDest>

export default meta
type Story = StoryObj<typeof meta>

export const EmptyList: Story = {
  args: {
    scheduleID: emptyScheduleID,
  },
}

export const GraphQLError: Story = {
  args: {
    scheduleID: errorScheduleID,
  },
}
export const MultipleConfigured: Story = {
  args: {
    scheduleID: manyNotificationsScheduleID,
  },
}
