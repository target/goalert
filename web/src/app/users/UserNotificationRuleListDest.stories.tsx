import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserNotificationRuleList from './UserNotificationRuleList'
import { expect, within, userEvent, screen } from '@storybook/test'
import { handleDefaultConfig } from '../storybook/graphql'
import { HttpResponse, graphql } from 'msw'

const meta = {
  title: 'users/UserNotificationRuleList',
  component: UserNotificationRuleList,
  tags: ['autodocs'],
  parameters: {
    msw: {
      handlers: [
        handleDefaultConfig,
        graphql.query('nrList', ({ variables: vars }) => {
          return HttpResponse.json({
            data:
              vars.id === '00000000-0000-0000-0000-000000000000'
                ? {
                    user: {
                      id: '00000000-0000-0000-0000-000000000000',
                      contactMethods: [
                        {
                          id: '12345',
                        },
                      ],
                      notificationRules: [
                        {
                          id: '123',
                          delayMinutes: 33,
                          contactMethod: {
                            id: '12345',
                            name: 'Josiah',
                            dest: {
                              type: 'single-field',
                              displayInfo: {
                                text: '+1 555-555-5555',
                                iconAltText: 'Voice Call',
                              },
                            },
                          },
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
                        },
                        {
                          id: '1111',
                        },
                      ],
                      notificationRules: [
                        {
                          id: '96814869-1199-4477-832d-e714e7d94aea',
                          delayMinutes: 71,
                          contactMethod: {
                            id: '67890',
                            name: 'Bridget',
                            dest: {
                              type: 'builtin-twilio-voice',
                              displayInfo: {
                                text: '+1 763-351-1103',
                                iconAltText: 'Voice Call',
                              },
                            },
                          },
                        },
                        {
                          id: 'eea77488-3748-4af8-99ba-18855f9a540d',
                          delayMinutes: 247,
                          contactMethod: {
                            id: '1111',
                            name: 'Dewayne',
                            dest: {
                              type: 'builtin-twilio-sms',
                              displayInfo: {
                                text: '+1 763-346-2643',
                                iconAltText: 'Text Message',
                              },
                            },
                          },
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
    return <UserNotificationRuleList {...args} />
  },
} satisfies Meta<typeof UserNotificationRuleList>

export default meta
type Story = StoryObj<typeof meta>

export const SingleContactMethod: Story = {
  args: {
    userID: '00000000-0000-0000-0000-000000000000',
    readOnly: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure correct info is displayed for single-field NR
    await expect(
      await canvas.findByText(
        'After 33 minutes notify me via Voice Call at +1 555-555-5555 (Josiah)',
      ),
    ).toBeVisible()

    await expect(await screen.findByText('Add Rule')).toHaveAttribute(
      'type',
      'button',
    )
    await userEvent.click(
      await screen.findByLabelText('Delete notification rule'),
    )
    await userEvent.click(await screen.findByText('Cancel'))
  },
}

export const MultiContactMethods: Story = {
  args: {
    userID: '00000000-0000-0000-0000-000000000001',
    readOnly: false,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    await expect(
      await canvas.findByText(
        'After 71 minutes notify me via Voice Call at +1 763-351-1103 (Bridget)',
      ),
    ).toBeVisible()
    await expect(
      await canvas.findByText(
        'After 247 minutes notify me via Text Message at +1 763-346-2643 (Dewayne)',
      ),
    ).toBeVisible()

    await expect(await screen.findByText('Add Rule')).toHaveAttribute(
      'type',
      'button',
    )
    const deleteButtons = await screen.findAllByLabelText(
      'Delete notification rule',
    )
    expect(deleteButtons).toHaveLength(2)
    await userEvent.click(deleteButtons[0])
    await userEvent.click(await screen.findByText('Cancel'))
  },
}

export const SingleReadOnlyContactMethods: Story = {
  args: {
    userID: '00000000-0000-0000-0000-000000000000',
    readOnly: true,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure no edit icons exist for read-only NR
    await expect(
      await canvas.queryByLabelText('Delete notification rule'),
    ).not.toBeInTheDocument()
  },
}
