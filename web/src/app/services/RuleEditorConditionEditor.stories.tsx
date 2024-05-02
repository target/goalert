import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { handleDefaultConfig, handleExpFlags } from '../storybook/graphql'
import { fn, userEvent, within, expect } from '@storybook/test'
import RuleEditorConditionEditor from './RuleEditorConditionEditor'
import { useArgs } from '@storybook/preview-api'

interface ClauseInput {
  field: string
  negate: boolean
  operator: string
  value: string
}

interface ConditionInput {
  clauses: ClauseInput[]
}

const meta = {
  title: 'Components/ConditionEditor',
  component: RuleEditorConditionEditor,
  args: {
    onChange: fn(),
  },
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onChange = (newValue: ConditionInput): void => {
      if (args.onChange) args.onChange(newValue)
      setArgs({ value: newValue })
    }
    return <RuleEditorConditionEditor {...args} onChange={onChange} />
  },
  tags: ['autodocs'],
  parameters: {
    docs: {
      story: {
        inline: false,
        iframeHeight: 500,
      },
    },
    msw: {
      handlers: [handleDefaultConfig, handleExpFlags('dest-types')],
    },
  },
} satisfies Meta<typeof RuleEditorConditionEditor>

export default meta
type Story = StoryObj<typeof meta>

const input: ConditionInput = {
  clauses: [{ field: 'foo', negate: false, operator: '==', value: '"bar"' }],
}

export const InputFieldText: Story = {
  args: {
    value: input,
    errors: null,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.click(await canvas.findByLabelText('Key'))

    const key = await canvas.findByLabelText('Key')
    expect(key).toHaveValue('foo')

    await expect(await canvas.findByText('==')).toBeVisible()

    await userEvent.click(await canvas.findByLabelText('Value'))

    const value = await canvas.findByLabelText('Value')
    expect(value).toHaveValue('bar')

    await expect(await canvas.findByText('string')).toBeVisible()
  },
}
