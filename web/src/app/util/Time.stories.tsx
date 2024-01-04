import type { Meta, StoryObj } from '@storybook/react'
import { Time } from './Time'
import { within, expect } from '@storybook/test'

const meta = {
  title: 'util/Time',
  component: Time,
  argTypes: {
    time: {
      control: 'date',
      defaultValue: Date.now(),
    },
    zone: {
      control: 'radio',
      options: ['utc', 'local', 'America/New_York', 'America/Los_Angeles'],
      defaultValue: 'local',
      if: { arg: 'time' },
    },
    format: {
      if: { arg: 'time' },
      defaultValue: 'default',
    },
    precise: {
      if: { arg: 'format', eq: 'relative' },
      defaultValue: false,
    },
    units: {
      control: 'multi-select',
      options: [
        'years',
        'months',
        'weeks',
        'days',
        'hours',
        'minutes',
        'seconds',
      ],
      defaultValue: ['days', 'hours', 'minutes'],
    },
  },
  decorators: [],
  tags: ['autodocs'],
} satisfies Meta<typeof Time>

export default meta
type Story = StoryObj<typeof meta>

export const CurrentTime: Story = {
  args: {
    time: new Date().toISOString(),
  },
}

export const DifferentZone: Story = {
  args: {
    time: new Date().toISOString(),
    zone: 'utc',
    format: 'default',
  },
}

const relTime = new Date(Date.now() - 3600000).toISOString()
export const RelativeTime: Story = {
  args: {
    // 1 hour ago
    time: relTime,
    zone: 'local',
    format: 'relative',
  },

  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(canvas.getByText('1 hr ago').getAttribute('title')).toContain(
      'local time',
    )
  },
}

export const RelativeTimePrecise: Story = {
  args: {
    // 1 hour ago
    time: new Date().toISOString(),
    zone: 'local',
    format: 'relative',
    precise: true,
    units: ['hours', 'minutes', 'seconds'],
  },
}

export const Duration: Story = {
  args: {
    // 1 hour ago
    duration: 'P1DT1H',
  },
}
