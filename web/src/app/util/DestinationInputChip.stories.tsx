import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import DestinationInputChip from './DestinationInputChip'
import { expect } from '@storybook/jest'
import { within } from '@storybook/testing-library'
import { handleDefaultConfig } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'

const meta = {
  title: 'util/DestinationInputChip',
  component: DestinationInputChip,
  render: function Component(args) {
    return <DestinationInputChip {...args} />
  },
  tags: ['autodocs'],
  parameters: {
    msw: {
      handlers: [
        handleDefaultConfig,
        graphql.query('DestDisplayInfo', () => {
          return HttpResponse.json({
            data: {
              destinationDisplayInfo: {
                text: 'Corporate array Communications Rotation',
                iconAltText: 'Rotation',
                iconURL: 'builtin://rotation',
                linkURL: 'test.com',
              },
            },
          })
        }),
      ],
    },
  },
} satisfies Meta<typeof DestinationInputChip>

export default meta
type Story = StoryObj<typeof meta>

export const Render: Story = {
  args: {
    value: {
      type: 'builtin-rotation',
      values: [
        {
          fieldID: 'rotation-id',
          value: 'bf227047-18b8-4de3-881c-24b9dd345670',
        },
      ],
    },
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    await expect(
      canvas.getByText('Corporate array Communications Rotation'),
    ).toBeVisible()

    await expect(canvas.getByTestId('destination-chip')).toHaveAttribute(
      'href',
      'test.com',
    )

    await expect(await canvas.findByTestId('RotateRightIcon')).toBeVisible()
    await expect(await canvas.findByTestId('CancelIcon')).toBeVisible()
  },
}
