import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import DestinationSearchSelect from './DestinationSearchSelect'
import { expect } from '@storybook/jest'
import { within } from '@storybook/testing-library'
import { handleDefaultConfig } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'

const meta = {
  title: 'util/DestinationSearchSelect',
  component: DestinationSearchSelect,
  render: function Component(args) {
    return <DestinationSearchSelect {...args} />
  },
  tags: ['autodocs'],
  parameters: {
    msw: {
      handlers: [
        handleDefaultConfig,
        graphql.query('DestinationSearchSelect', () => {
          return HttpResponse.json({
            data: {
              destinationFieldSearch: {
                nodes: [
                  {
                    value: 'C03SJES5FA7',
                    label: '#general',
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
          return HttpResponse.json({
            data: {
              destinationFieldValueName:
                vars.input.value === 'C03SJES5FA7' ? '#general' : '',
            },
          })
        }),
      ],
    },
  },
} satisfies Meta<typeof DestinationSearchSelect>

export default meta
type Story = StoryObj<typeof meta>

export const Render: Story = {
  args: {
    value: '',
    config: {
      fieldID: 'slack-channel-id',
      hint: '',
      hintURL: '',
      inputType: 'text',
      isSearchSelectable: true,
      labelPlural: 'Slack Channels',
      labelSingular: 'Slack Channel',
      placeholderText: '',
      prefix: '',
      supportsValidation: false,
    },
    destType: 'builtin-slack-channel',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // should see #general channel as option
    await userEvent.type(
      canvas.getByPlaceholderText('Start typing...'),
      '#general',
    )
  },
}
