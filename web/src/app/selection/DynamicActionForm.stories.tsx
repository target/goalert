import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'

import DynamicActionForm, { Value } from './DynamicActionForm'
import { DestinationTypeInfo } from '../../schema'
import { fn } from '@storybook/test'
import { useArgs } from '@storybook/preview-api'

const meta = {
  title: 'util/Destination/DynamicActionForm',
  component: DynamicActionForm,
  args: {
    onChange: fn(),
  },
  parameters: {
    graphql: {
      RequireConfig: {
        data: {
          destinationTypes: [
            {
              type: 'action-type',
              name: 'Action Type',
              enabled: true,
              userDisclaimer: '',
              isContactMethod: false,
              isEPTarget: false,
              isSchedOnCallNotify: false,
              isDynamicAction: true,
              iconURL: '',
              iconAltText: '',
              supportsStatusUpdates: false,
              statusUpdatesRequired: false,
              dynamicParams: [
                {
                  paramID: 'dynamic-param',
                  label: 'Dynamic Param',
                  hint: 'Param Hint',
                  hintURL: 'http://example.com/hint',
                  defaultValue: '',
                },
              ],
              requiredFields: [
                {
                  fieldID: 'required-field',
                  label: 'Required Field',
                  hint: 'Field Hint',
                  hintURL: '',
                  placeholderText: '',
                  prefix: '',
                  inputType: 'text',
                  supportsSearch: false,
                  supportsValidation: false,
                },
              ],
            },
            {
              type: 'action-type-2',
              name: 'Action Type 2',
              enabled: true,
              userDisclaimer: '',
              isContactMethod: false,
              isEPTarget: false,
              isSchedOnCallNotify: false,
              isDynamicAction: true,
              iconURL: '',
              iconAltText: '',
              supportsStatusUpdates: false,
              statusUpdatesRequired: false,
              dynamicParams: [
                {
                  paramID: 'dynamic-param-2',
                  label: 'Dynamic Param 2',
                  hint: 'Param Hint 2',
                  hintURL: 'http://example.com/hint2',
                  defaultValue: '',
                },
              ],
              requiredFields: [
                {
                  fieldID: 'required-field-2',
                  label: 'Required Field 2',
                  hint: 'Field Hint 2',
                  hintURL: '',
                  placeholderText: '',
                  prefix: '',
                  inputType: 'text',
                  supportsSearch: false,
                  supportsValidation: false,
                },
              ],
            },
          ] satisfies DestinationTypeInfo[],
        },
      },
    },
  },
  tags: ['autodocs'],
} satisfies Meta<typeof DynamicActionForm>

export default meta
type Story = StoryObj<typeof meta>

export const Interactive: Story = {
  args: {
    value: null,
  },
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onChange = (value: Value): void => {
      if (args.onChange) args.onChange(value)
      setArgs({ value })
    }
    return <DynamicActionForm {...args} onChange={onChange} />
  },
}

export const NullValue: Story = {
  args: {
    value: null,
  },
}

export const Values: Story = {
  args: {
    value: {
      destType: 'action-type',
      staticParams: { 'required-field': 'value' },
      dynamicParams: { 'dynamic-param': 'req.body.dynamic-param' },
    },
  },
}

export const Errors: Story = {
  args: {
    value: {
      destType: 'action-type',
      staticParams: {},
      dynamicParams: {},
    },
    destTypeError: 'Dest type validation error',
    staticParamErrors: { 'required-field': 'Required field validation error' },
    dynamicParamErrors: { 'dynamic-param': 'Dynamic param validation error' },
  },
}
