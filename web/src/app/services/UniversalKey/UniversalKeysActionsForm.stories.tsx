import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { expect, userEvent, waitFor, within, fn } from '@storybook/test'
import { useArgs } from '@storybook/preview-api'
import UniversalKeyActionsForm from './UniversalKeyActionsForm'
import { ActionInput, DestinationDisplayInfo } from '../../../schema'

const meta = {
  title: 'UIK/Actions Form',
  component: UniversalKeyActionsForm,
  args: {
    onChange: fn(),
  },
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onChange = (newValue: Array<ActionInput>): void => {
      if (args.onChange) args.onChange(newValue)
      setArgs({ value: newValue })
    }
    return (
      <UniversalKeyActionsForm {...args} onChange={onChange} disablePortal />
    )
  },
  tags: ['autodocs'],
} satisfies Meta<typeof UniversalKeyActionsForm>

export default meta
type Story = StoryObj<typeof meta>
export const Empty: Story = {
  args: {
    value: [],
  },
}

export const WithList: Story = {
  args: {
    showList: true,
    value: [{ dest: { type: 'foo', args: {} }, params: {} }],
  },
  parameters: {
    graphql: {
      DestDisplayInfo: {
        data: {
          destinationDisplayInfo: {
            text: 'VALID_CHIP_1',
            iconURL: 'builtin://rotation',
            linkURL: '',
            iconAltText: 'Rotation',
          } satisfies DestinationDisplayInfo,
        },
      },
    },
  },
}

export const ValidationError: Story = {
  args: {
    showList: true,
    value: [],
  },
  parameters: {
    graphql: {
      ValidateActionInput: {
        errors: [
          { message: 'generic error' },
          {
            path: ['actionInputValidate', 'input', 'dest', 'type'],
            message: 'invalid type',
          },
          {
            path: ['actionInputValidate', 'input', 'dest', 'args'],
            extensions: {
              key: 'phone_number',
            },
            message: 'invalid number',
          },
          {
            path: ['actionInputValidate', 'input', 'params'],
            extensions: {
              key: 'example-param',
            },
            message: 'invalid param',
          },
        ],
      },
    },
  },

  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    userEvent.click(canvas.getByLabelText('Destination Type'))
    userEvent.click(await canvas.findByText('Single Field'))
    await canvas.findByLabelText('Phone Number')
    userEvent.click(await canvas.findByRole('button', { name: /add/i }))

    waitFor(async () => {
      expect(await canvas.getByLabelText('Phone Number')).toBeInvalid()
      expect(
        await canvas.getByLabelText('Example Param (Expr syntax)'),
      ).toBeInvalid()
      expect(canvas.getByText('generic error')).toBeVisible()
      expect(canvas.getByText('invalid type')).toBeVisible()
    })
  },
}
