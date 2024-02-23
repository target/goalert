import type { Meta, StoryObj } from '@storybook/react'
import ScheduleOnCallNotificationsFormDest from './ScheduleOnCallNotificationsFormDest'

const meta = {
  title: 'schedules/on-call-notifications/FormDest',
  component: ScheduleOnCallNotificationsFormDest,
  argTypes: {},
  args: {
    scheduleID: '',
    value: {
      time: null,
      weekdayFilter: [false, false, false, false, false, false, false],
      dest: {
        type: 'single-field',
        values: [],
      },
    },
  },
  tags: ['autodocs'],
} satisfies Meta<typeof ScheduleOnCallNotificationsFormDest>

export default meta
type Story = StoryObj<typeof meta>

export const Empty: Story = {
  args: {},
}
