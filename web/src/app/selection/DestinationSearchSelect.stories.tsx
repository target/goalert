import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import DestinationSearchSelect from './DestinationSearchSelect'
import { expect, userEvent, within } from '@storybook/test'
import { handleDefaultConfig } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { useArgs } from '@storybook/preview-api'
import { FieldValueConnection } from '../../schema'

const meta = {
  title: 'util/DestinationSearchSelect',
  component: DestinationSearchSelect,
  argTypes: {
    inputType: { table: { disable: true } },
    placeholderText: { table: { disable: true } },
    isSearchSelectable: { table: { disable: true } },
    supportsValidation: { table: { disable: true } },
    prefix: { table: { disable: true } },
  },
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
        graphql.query('DestinationFieldSearch', ({ variables: vars }) => {
          if (vars.input.search === 'query-error') {
            return HttpResponse.json({
              errors: [{ message: 'some_backend_error_message' }],
            })
          }
          if (vars.input.search === 'empty') {
            return HttpResponse.json({
              data: {
                destinationFieldSearch: {
                  nodes: [],
                  __typename: 'FieldValueConnection',
                },
              },
            })
          }

          return HttpResponse.json({
            data: {
              destinationFieldSearch: {
                nodes: [
                  {
                    fieldID: 'field-id',
                    value: 'value-id-1',
                    label: '#value-one',
                    isFavorite: false,
                  },
                  {
                    fieldID: 'field-id',
                    value: 'value-id-2',
                    label: '#value-two',
                    isFavorite: false,
                  },
                ],
              },
            },
          } satisfies {
            data: { destinationFieldSearch: Partial<FieldValueConnection> }
          })
        }),
        graphql.query('DestinationFieldValueName', ({ variables: vars }) => {
          if (vars.input.value === 'invalid-value') {
            return HttpResponse.json({
              errors: [{ message: 'some_backend_error_message' }],
            })
          }

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

export const Default: Story = {
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
}

export const OptionSelected: Story = {
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
}

export const InvalidOptionSelected: Story = {
  args: {
    value: 'invalid-value',

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
}

export const NoOptions: Story = {
  args: {
    value: '',

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
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    const box = await canvas.findByRole('combobox')
    await userEvent.click(box)
    await userEvent.type(box, 'empty', { delay: null })

    expect(await within(document.body).findByText('No options')).toBeVisible()
  },
}

export const QueryError: Story = {
  args: {
    value: '',

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
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    const box = await canvas.findByRole('combobox')
    await userEvent.click(box)
    await userEvent.type(box, 'query-error', { delay: null })

    expect(
      await within(document.body).findByText(
        '[GraphQL] some_backend_error_message',
      ),
    ).toBeVisible()
  },
}
