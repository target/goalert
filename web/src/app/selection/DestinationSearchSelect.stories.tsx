import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import DestinationSearchSelect from './DestinationSearchSelect'
import { expect, userEvent, within } from '@storybook/test'
import { mockOp } from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'
import {
  DestinationFieldSearchInput,
  FieldSearchConnection,
  FieldValueInput,
} from '../../schema'

const meta = {
  title: 'util/DestinationSearchSelect',
  component: DestinationSearchSelect,
  argTypes: {
    inputType: { table: { disable: true } },
    placeholderText: { table: { disable: true } },
    supportsSearch: { table: { disable: true } },
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
    fetchMock: {
      mocks: [
        mockOp<DestinationFieldSearchInput>(
          'DestinationFieldSearch',
          (vars) => {
            if (vars.input.search === 'query-error') {
              return {
                errors: [{ message: 'some_backend_error_message' }],
              }
            }
            if (vars.input.search === 'empty') {
              return {
                data: {
                  destinationFieldSearch: {
                    nodes: [],
                    __typename: 'FieldSearchConnection',
                  },
                },
              }
            }

            return {
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
              data: { destinationFieldSearch: Partial<FieldSearchConnection> }
            }
          },
        ),
        mockOp<FieldValueInput>('DestinationFieldValueName', (vars) => {
          if (vars.input.value === 'invalid-value') {
            return {
              errors: [{ message: 'some_backend_error_message' }],
            }
          }

          const names: Record<string, string> = {
            'value-id-1': '#value-one',
            'value-id-2': '#value-two',
          }
          return {
            data: {
              destinationFieldValueName: names[vars.input.value] || '',
            },
          }
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
    supportsSearch: true,
    label: 'Select Value',
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
    supportsSearch: true,
    label: 'Select Value',
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
    supportsSearch: true,
    label: 'Select Value',
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
    supportsSearch: true,
    label: 'Select Value',
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
    supportsSearch: true,
    label: 'Select Value',
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
    supportsSearch: true,
    label: 'Select Value',
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
