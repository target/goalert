import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import ScheduleOnCallNotificationsForm, {
  Value,
} from './ScheduleOnCallNotificationsForm'
import { useArgs } from '@storybook/preview-api'
import { expect, fn, userEvent, within } from '@storybook/test'

const meta = {
  title: 'schedules/on-call-notifications/FormDest',
  component: ScheduleOnCallNotificationsForm,
  argTypes: {},
  args: {
    scheduleID: '',
    onChange: fn(),
    value: {
      time: null,
      weekdayFilter: [false, false, false, false, false, false, false],
      dest: {
        type: 'single-field',
        args: {},
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
    return <ScheduleOnCallNotificationsForm {...args} onChange={onChange} />
  },
} satisfies Meta<typeof ScheduleOnCallNotificationsForm>

export default meta
type Story = StoryObj<typeof meta>

export const Empty: Story = {
  args: {},
}

export const ValidationErrors: Story = {
  args: {
    destTypeError: 'error with dest type',
    destFieldErrors: {
      phone_number: 'error with dest field',
    },
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

    await expect(await canvas.findByText('Error with dest field')).toBeVisible()
    await expect(await canvas.findByText('Error with dest type')).toBeVisible()
  },
}
