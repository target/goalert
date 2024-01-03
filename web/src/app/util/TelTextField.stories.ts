import type { Meta, StoryObj } from '@storybook/react'
import TelTextField from './TelTextField'
import { graphql } from 'msw'

const meta = {
  title: 'util/TelTextField',
  component: TelTextField,
  argTypes: {
    component: { table: { disable: true } },
    ref: { table: { disable: true } },

    value: { control: 'text', defaultValue: '+17635550123' },
    label: { control: 'text', defaultValue: 'Phone Number' },
    error: { control: 'boolean' },
    onChange: { action: 'onChange' },
  },
  decorators: [],
  tags: ['autodocs'],
} satisfies Meta<typeof TelTextField>

export default meta
type Story = StoryObj<typeof meta>

export const ValidNumber: Story = {
  //   msw: {
  //     handlers: [
  //       graphql.query('AllFilmsQuery', (req, res, ctx) => {
  //         return res(
  //           ctx.delay(800),
  //           ctx.errors([
  //             {
  //               message: 'Access denied',
  //             },
  //           ]),
  //         )
  //       }),
  //     ],
  //   },
  args: {
    value: '+17635550123',
    label: 'Phone Number',
    error: false,
  },
}
