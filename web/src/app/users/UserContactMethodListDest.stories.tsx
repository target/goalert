import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodListDest from './UserContactMethodListDest'
import { expect, within, userEvent, screen } from '@storybook/test'
import { handleDefaultConfig } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'
import { Destination } from '../../schema'

const meta = {
  title: 'users/UserContactMethodListDest',
  component: UserContactMethodListDest,
  tags: ['autodocs'],
  parameters: {
    msw: {
      handlers: [
        handleDefaultConfig,
        graphql.query('cmList', ({ variables: vars }) => {
          return HttpResponse.json({
            data:
              vars.id === '00000000-0000-0000-0000-000000000000'
                ? {
                    user: {
                      id: '00000000-0000-0000-0000-000000000000',
                      contactMethods: [
                        {
                          id: '12345',
                          name: 'Josiah',
                          dest: {
                            type: 'single-field',
                            values: [
                              {
                                fieldID: 'phone-number',
                                value: '+15555555555',
                              },
                            ],
                            displayInfo: {
                              text: '+1 555-555-5555',
                              iconAltText: 'Voice Call',
                              iconURL: '',
                              linkURL: '',
                            },
                          } satisfies Destination,
                          disabled: false,
                          pending: false,
                        },
                      ],
                    },
                  }
                : {
                    user: {
                      id: '00000000-0000-0000-0000-000000000001',
                      contactMethods: [
                        {
                          id: '67890',
                          name: 'triple contact method',
                          dest: {
                            type: 'triple-field',
                            values: [
                              {
                                fieldID: 'first-field',
                                value: 'test_user@target.com',
                              },
                              {
                                fieldID: 'second-field',
                                value: 'parameter-1',
                              },
                              {
                                fieldID: 'third-field',
                                value: 'parameter-2',
                              },
                            ],
                            displayInfo: {
                              text: 'test_user@target.com',
                              iconAltText: 'Email',
                              iconURL: '',
                              linkURL: '',
                            },
                          } satisfies Destination,
                          disabled: false,
                          pending: false,
                        },
                        {
                          id: '1111',
                          name: 'single field CM',
                          dest: {
                            type: 'single-field',
                            values: [
                              {
                                fieldID: 'phone-number',
                                value: '+15555555556',
                              },
                            ],
                            displayInfo: {
                              text: '+1 555-555-5556',
                              iconAltText: 'Voice Call',
                              iconURL: '',
                              linkURL: '',
                            },
                          } satisfies Destination,
                          disabled: false,
                          pending: false,
                        },
                      ],
                    },
                  },
          })
        }),
      ],
    },
  },
  render: function Component(args) {
    return <UserContactMethodListDest {...args} />
  },
} satisfies Meta<typeof UserContactMethodListDest>

export default meta
type Story = StoryObj<typeof meta>

export const SingleContactMethod: Story = {
  args: {
    userID: '00000000-0000-0000-0000-000000000000',
    readOnly: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure correct info is displayed for single-field CM
    await expect(await canvas.findByText('Josiah (Single Field)')).toBeVisible()
    await expect(await canvas.findByText('+1 555-555-5555')).toBeVisible()
    // ensure CM is editable
    await expect(
      await canvas.queryByTestId('MoreHorizIcon'),
    ).toBeInTheDocument()
    // ensure all edit options are available
    await userEvent.click(await canvas.findByTestId('MoreHorizIcon'))
    await expect(await screen.findByText('Edit')).toHaveAttribute(
      'role',
      'menuitem',
    )
    await expect(await screen.findByText('Delete')).toHaveAttribute(
      'role',
      'menuitem',
    )
    await expect(await screen.findByText('Send Test')).toHaveAttribute(
      'role',
      'menuitem',
    )
  },
}

export const MultiContactMethods: Story = {
  args: {
    userID: '00000000-0000-0000-0000-000000000001',
    readOnly: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure correct info is displayed for single field CM
    await expect(
      await canvas.findByText('single field CM (Single Field)'),
    ).toBeVisible()
    // ensure correct info is displayed for triple-field CM
    await expect(
      await canvas.findByText('triple contact method (Multi Field)'),
    ).toBeVisible()
    await expect(await canvas.findByText('+1 555-555-5556')).toBeVisible()
    await expect(await canvas.findByText('test_user@target.com')).toBeVisible()
    // ensure all edit icons exists
    await expect(await canvas.findAllByTestId('MoreHorizIcon')).toHaveLength(2)
  },
}

export const SingleReadOnlyContactMethods: Story = {
  args: {
    userID: '00000000-0000-0000-0000-000000000000',
    readOnly: true,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure no edit icons exist for read-only CM
    await expect(
      await canvas.queryByTestId('MoreHorizIcon'),
    ).not.toBeInTheDocument()
  },
}
