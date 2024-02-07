import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodEditDialogDest from './UserContactMethodEditDialogDest'
import { expect } from '@storybook/jest'
import { screen, userEvent, waitFor } from '@storybook/testing-library'
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
    // setup
    const destTypeOptions = await screen.getByText(
      'Single Field Destination Type',
    )
    const phoneNumInput = await screen.getByDisplayValue('11235555555')
    const nameLabel = screen.getByLabelText('Name')

    // ensure correct values are displayed
    await waitFor(async () => {
      await expect(nameLabel).toBeVisible()
      await expect(screen.getByLabelText('Dest Type')).toBeVisible()
      await expect(screen.getByLabelText('Phone Number')).toBeVisible()
    })
    await expect(screen.getByDisplayValue('test_cm')).toBeVisible()
    await expect(phoneNumInput).toBeVisible()
    await expect(destTypeOptions).toBeVisible()

    // ensure dest-type, phone number, and alert status are disabled
    userEvent.click(destTypeOptions, {
      pointerEventsCheck: 0,
    })
    userEvent.click(phoneNumInput, {
      pointerEventsCheck: 0,
    })

    const status = await screen.getByLabelText(
      'Send alert status updates (not supported for this type)',
    )
    userEvent.click(status, {
      pointerEventsCheck: 0,
    })

    // ensure we can update name and submit
    await userEvent.clear(nameLabel)
    await userEvent.type(nameLabel, 'changed')

    const submitButton = await screen.getByRole('button', { name: /SUBMIT/i })
    await userEvent.click(submitButton)
  },
}

export const MultiField: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000002',
  },
  play: async () => {
    // setup
    const destTypeOptions = await screen.getByText(
      'Multi Field Destination Type',
    )
    const firstField = await screen.getByDisplayValue('12225559999')
    const secondField = await screen.getByDisplayValue('multiemail@target.com')
    const thirdField = await screen.getByDisplayValue('slackID')
    const nameLabel = screen.getByLabelText('Name')

    // ensure correct values are displayed for all fields
    await waitFor(async () => {
      await expect(nameLabel).toBeVisible()
      await expect(screen.getByLabelText('Dest Type')).toBeVisible()
      await expect(screen.getByLabelText('First Item')).toBeVisible()
      await expect(screen.getByLabelText('Second Item')).toBeVisible()
      await expect(screen.getByLabelText('Third Item')).toBeVisible()
    })
    await expect(screen.getByDisplayValue('test_cm')).toBeVisible()
    await expect(firstField).toBeVisible()
    await expect(destTypeOptions).toBeVisible()
    await expect(secondField).toBeVisible()
    await expect(thirdField).toBeVisible()

    // ensure dest-type, all fields, and alert status are disabled
    userEvent.click(destTypeOptions, {
      pointerEventsCheck: 0,
    })
    userEvent.click(firstField, {
      pointerEventsCheck: 0,
    })
    userEvent.click(secondField, {
      pointerEventsCheck: 0,
    })
    userEvent.click(thirdField, {
      pointerEventsCheck: 0,
    })

    const status = await screen.getByLabelText(
      'Send alert status updates (not supported for this type)',
    )
    userEvent.click(status, {
      pointerEventsCheck: 0,
    })

    // ensure we can update name and submit
    await userEvent.clear(nameLabel)
    await userEvent.type(nameLabel, 'changed')

    const submitButton = await screen.getByRole('button', { name: /SUBMIT/i })
    await userEvent.click(submitButton)
  },
}
