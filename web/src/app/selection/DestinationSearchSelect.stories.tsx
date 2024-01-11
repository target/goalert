import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import DestinationSearchSelect from './DestinationSearchSelect'
import { expect } from '@storybook/jest'
import { within } from '@storybook/testing-library'
import { handleDefaultConfig } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { useArgs } from '@storybook/preview-api'

const meta = {
  title: 'util/DestinationSearchSelect',
  component: DestinationSearchSelect,
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onChange = (value: string): void => {
      if (args.onChange) args.onChange(value)
      setArgs({ value })
    }
    return <DestinationSearchSelect {...args} onChange={onChange} />
  },
  tags: ['autodocs'],
  parameters: {
    msw: {
      handlers: [
        handleDefaultConfig,
        graphql.query('DestinationFieldSearch', () => {
          return HttpResponse.json({
            data: {
              destinationFieldSearch: {
                nodes: [
                  {
                    value: 'value-id-1',
                    label: '#value-one',
                    isFavorite: false,
                    __typename: 'FieldValuePair',
                  },
                  {
                    value: 'value-id-2',
                    label: '#value-two',
                    isFavorite: false,
                    __typename: 'FieldValuePair',
                  },
                ],
                __typename: 'FieldValueConnection',
              },
            },
          })
        }),
        graphql.query('DestinationFieldValueName', ({ variables: vars }) => {
          const names: Record<string, string> = {
            'value-id-1': '#value-one',
            'value-id-2': '#value-two',
          }
          return HttpResponse.json({
            data: {
              destinationFieldValueName: names[vars.input.value] || '',
            },
          })
        }),
      ],
    },
  },
} satisfies Meta<typeof DestinationSearchSelect>

export default meta
type Story = StoryObj<typeof meta>

export const EmptyValue: Story = {
  args: {
    value: '',

    fieldID: 'field-id',
    hint: '',
    hintURL: '',
    inputType: 'text',
    isSearchSelectable: true,
    labelPlural: 'Select Values',
    labelSingular: 'Select Value',
    placeholderText: 'asdf',
    prefix: '',
    supportsValidation: false,

    destType: 'test-type',
  },
  // play: async ({ canvasElement }) => {
  //   const canvas = within(canvasElement)

  //   // should see #general channel as option
  //   await userEvent.type(
  //     canvas.getByPlaceholderText('Start typing...'),
  //     '#general',
  //   )
  // },
}

export const SelectedValue: Story = {
  args: {
    value: 'value-id-1',

    fieldID: 'field-id',
    hint: '',
    hintURL: '',
    inputType: 'text',
    isSearchSelectable: true,
    labelPlural: 'Select Values',
    labelSingular: 'Select Value',
    placeholderText: '',
    prefix: '',
    supportsValidation: false,

    destType: 'test-type',
  },
  // play: async ({ canvasElement }) => {
  //   const canvas = within(canvasElement)

  //   // should see #general channel as option
  //   await userEvent.type(
  //     canvas.getByPlaceholderText('Start typing...'),
  //     '#general',
  //   )
  // },
}

export const Disabled: Story = {
  args: {
    value: 'value-id-1',

    fieldID: 'field-id',
    hint: '',
    hintURL: '',
    inputType: 'text',
    isSearchSelectable: true,
    labelPlural: 'Select Values',
    labelSingular: 'Select Value',
    placeholderText: '',
    prefix: '',
    supportsValidation: false,

    destType: 'test-type',
    disabled: true,
  },
  // play: async ({ canvasElement }) => {
  //   const canvas = within(canvasElement)

  //   // should see #general channel as option
  //   await userEvent.type(
  //     canvas.getByPlaceholderText('Start typing...'),
  //     '#general',
  //   )
  // },
}
