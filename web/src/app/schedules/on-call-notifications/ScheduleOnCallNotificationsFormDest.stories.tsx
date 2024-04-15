import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import ScheduleOnCallNotificationsFormDest, {
  Value,
} from './ScheduleOnCallNotificationsFormDest'
import { useArgs } from '@storybook/preview-api'
import { expect, fn, userEvent, within } from '@storybook/test'

const meta = {
  title: 'schedules/on-call-notifications/FormDest',
  component: ScheduleOnCallNotificationsFormDest,
  argTypes: {},
  args: {
    scheduleID: '',
    onChange: fn(),
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
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onChange = (newValue: Value): void => {
      if (args.onChange) args.onChange(newValue)
      setArgs({ value: newValue })
    }
    return <ScheduleOnCallNotificationsFormDest {...args} onChange={onChange} />
  },
} satisfies Meta<typeof ScheduleOnCallNotificationsFormDest>

export default meta
type Story = StoryObj<typeof meta>

export const Empty: Story = {
  args: {},
}

export const ValidationErrors: Story = {
  args: {
    errors: [
      {
        path: ['mutation', 'input', 'time'],
        message: 'error with time',
        extensions: {
          code: 'INVALID_INPUT_VALUE',
        },
      },
      {
        path: ['mutation', 'input', 'dest'],
        message: 'error with dest field',
        extensions: {
          code: 'INVALID_DEST_FIELD_VALUE',
          fieldID: 'phone-number',
        },
      },
      {
        path: ['mutation', 'input', 'dest', 'type'],
        message: 'error with dest type',
        extensions: {
          code: 'INVALID_INPUT_VALUE',
        },
      },
    ],
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    await userEvent.click(
      await canvas.findByLabelText(
        'Notify at a specific day and time every week',
      ),
    )

    await expect(await canvas.findByLabelText('Time')).toBeInvalid()
    await expect(
      // mui puts aria-invalid on the input, but not the combobox (which the label points to)
      canvasElement.querySelector('input[name="dest.type"]'),
    ).toBeInvalid()
    await expect(await canvas.findByLabelText('Phone Number')).toBeInvalid()

    await expect(await canvas.findByText('Error with time')).toBeVisible()
    await expect(await canvas.findByText('Error with dest field')).toBeVisible()
    await expect(await canvas.findByText('Error with dest type')).toBeVisible()
  },
}
