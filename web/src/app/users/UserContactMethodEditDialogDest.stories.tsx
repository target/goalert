import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodEditDialogDest from './UserContactMethodEditDialogDest'
import { expect, screen, userEvent, waitFor } from '@storybook/test'
import { handleDefaultConfig } from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'
import { HttpResponse, graphql } from 'msw'

const meta = {
  title: 'users/UserContactMethodEditDialogDest',
  component: UserContactMethodEditDialogDest,
  tags: ['autodocs'],
  parameters: {
    msw: {
      handlers: [
        handleDefaultConfig,
        graphql.query('UserContactMethod', ({ variables: vars }) => {
          return HttpResponse.json({
            data: {
              userContactMethod:
                vars.id === '00000000-0000-0000-0000-000000000001'
                  ? {
                      id: '00000000-0000-0000-0000-000000000001',
                      name: 'test_cm',
                      dest: {
                        type: 'single-field',
                        values: [
                          {
                            fieldID: 'phone-number',
                            value: '+11235555555',
                            __typename: 'FieldValuePair',
                          },
                        ],
                        __typename: 'Destination',
                      },
                      statusUpdates: 'DISABLED',
                      __typename: 'UserContactMethod',
                    }
                  : {
                      id: '00000000-0000-0000-0000-000000000002',
                      name: 'test_cm',
                      dest: {
                        type: 'triple-field',
                        values: [
                          {
                            fieldID: 'first-field',
                            value: '+12225559999',
                            __typename: 'FieldValuePair',
                          },
                          {
                            fieldID: 'second-field',
                            value: 'multiemail@target.com',
                            __typename: 'FieldValuePair',
                          },
                          {
                            fieldID: 'third-field',
                            value: 'slackID',
                            __typename: 'FieldValuePair',
                          },
                        ],
                        __typename: 'Destination',
                      },
                      statusUpdates: 'DISABLED',
                      __typename: 'UserContactMethod',
                    },
            },
          })
        }),
        graphql.mutation('UpdateUserContactMethod', () => {
          return HttpResponse.json({
            data: {
              updateUserContactMethod: true,
            },
          })
        }),
      ],
    },
  },
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onClose = (): void => {
      if (args.onClose) args.onClose()
      setArgs({ value: '' })
    }
    return <UserContactMethodEditDialogDest {...args} onClose={onClose} />
  },
} satisfies Meta<typeof UserContactMethodEditDialogDest>

export default meta
type Story = StoryObj<typeof meta>

export const SingleField: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000001',
  },
  play: async () => {
    // ensure correct values are displayed and disabled
    await waitFor(async () => {
      await expect(await screen.findByLabelText('Name')).toBeVisible()
      await expect(await screen.findByLabelText('Dest Type')).toHaveAttribute(
        'aria-disabled',
        'true',
      )
      await expect(
        await screen.findByPlaceholderText('11235550123'),
      ).toBeDisabled()
      await expect(
        await screen.findByLabelText(
          'Send alert status updates (not supported for this type)',
        ),
      ).toBeDisabled()
    })
  },
}

export const MultiField: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000002',
  },
  play: async () => {
    // ensure correct values are displayed and disabled for all fields
    await waitFor(async () => {
      await expect(await screen.findByLabelText('Name')).toBeVisible()
      await expect(await screen.findByLabelText('Dest Type')).toBeVisible()
      await expect(await screen.findByLabelText('First Item')).toBeVisible()
      await expect(await screen.findByLabelText('Second Item')).toBeVisible()
      await expect(await screen.findByLabelText('Third Item')).toBeVisible()

      await expect(await screen.findByLabelText('Dest Type')).toHaveAttribute(
        'aria-disabled',
        'true',
      )
      await expect(
        await screen.findByPlaceholderText('11235550123'),
      ).toBeDisabled()
      await expect(
        await screen.findByPlaceholderText('foobar@example.com'),
      ).toBeDisabled()
      await expect(
        await screen.findByPlaceholderText('slack user ID'),
      ).toBeDisabled()
    })

    await expect(
      await screen.findByLabelText('Send alert status updates'),
    ).not.toBeDisabled()

    // ensure we can update name and submit
    await userEvent.clear(await screen.findByLabelText('Name'))
    await userEvent.type(await screen.findByLabelText('Name'), 'changed')

    const submitButton = await screen.findByRole('button', { name: /SUBMIT/i })
    await userEvent.click(submitButton)
  },
}
