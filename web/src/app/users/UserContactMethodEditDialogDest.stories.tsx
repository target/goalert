import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodEditDialogDest from './UserContactMethodEditDialogDest'
import { expect, userEvent, waitFor, screen } from '@storybook/test'
import {
  handleDefaultConfig,
  // defaultConfig,
  handleExpFlags,
} from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'
import { HttpResponse, graphql } from 'msw'
import { DestFieldValueError, InputFieldError } from '../util/errtypes'

const meta = {
  title: 'users/UserContactMethodEditDialogDest',
  component: UserContactMethodEditDialogDest,
  tags: ['autodocs'],
  parameters: {
    docs: {
      story: {
        inline: false,
        iframeHeight: 500,
      },
    },
    msw: {
      handlers: [
        handleDefaultConfig,
        handleExpFlags('dest-types'),
        graphql.query('userCm', ({ variables: vars }) => {
          return HttpResponse.json({
            data: {
              userContactMethod:
                vars.id === '00000000-0000-0000-0000-000000000000'
                  ? {
                      id: '00000000-0000-0000-0000-000000000000',
                      name: 'single-field contact method',
                      dest: {
                        type: 'supports-status',
                        values: [
                          {
                            fieldID: 'phone-number',
                            value: '+15555555555',
                            label: '+1 555-555-5555',
                          },
                        ],
                      },
                      value: 'http://localhost:8080',
                      statusUpdates: 'DISABLED',
                      disabled: false,
                      pending: false,
                    }
                  : {
                      id: '00000000-0000-0000-0000-000000000001',
                      name: 'Multi Field',
                      dest: {
                        type: 'triple-field',
                        values: [
                          {
                            fieldID: 'first-field',
                            label: '+1 555-555-5555',
                            value: '+11235550123',
                          },
                          {
                            fieldID: 'second-field',
                            label: 'email',
                            value: 'foobar@example.com',
                          },
                          {
                            fieldID: 'third-field',
                            label: 'slack user ID',
                            value: 'slack',
                          },
                        ],
                      },
                      statusUpdates: 'ENABLED',
                      disabled: false,
                      pending: false,
                    },
            },
          })
        }),
        graphql.mutation('UpdateUserContactMethod', ({ variables: vars }) => {
          if (vars.input.name === 'error-test') {
            return HttpResponse.json({
              data: null,
              errors: [
                {
                  message: 'This is a dest field-error',
                  path: ['updateUserContactMethod', 'input', 'dest'],
                  extensions: {
                    code: 'INVALID_DEST_FIELD_VALUE',
                    fieldID: 'phone-number',
                  },
                } satisfies DestFieldValueError,
                {
                  message: 'This indicates an invalid destination type',
                  path: ['updateUserContactMethod', 'input', 'dest', 'type'],
                  extensions: {
                    code: 'INVALID_INPUT_VALUE',
                  },
                } satisfies InputFieldError,
                {
                  message: 'Name error',
                  path: ['updateUserContactMethod', 'input', 'name'],
                  extensions: {
                    code: 'INVALID_INPUT_VALUE',
                  },
                } satisfies InputFieldError,
                {
                  message: 'This is a generic error',
                },
              ],
            })
          }
          return HttpResponse.json({
            data: {
              updateUserContactMethod: {
                id: '00000000-0000-0000-0000-000000000000',
              },
            },
          })
        }),
        graphql.query('ValidateDestination', ({ variables: vars }) => {
          return HttpResponse.json({
            data: {
              destinationFieldValidate:
                vars.input.value === '@slack' ||
                vars.input.value === '+12225558989' ||
                vars.input.value === 'valid@email.com',
            },
          })
        }),
      ],
    },
  },
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onClose = (contactMethodID: string | undefined): void => {
      if (args.onClose) args.onClose(contactMethodID)
      setArgs({ value: contactMethodID })
    }
    return <UserContactMethodEditDialogDest {...args} onClose={onClose} />
  },
} satisfies Meta<typeof UserContactMethodEditDialogDest>

export default meta
type Story = StoryObj<typeof meta>

export const SingleField: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000000',
  },
  play: async () => {
    await userEvent.click(await screen.findByLabelText('Destination Type'))

    const [single] = await screen.findAllByRole('combobox')
    expect(single).toHaveTextContent('Single With Status')
    await screen.findByTestId('CheckBoxOutlineBlankIcon')
  },
}

export const MultiField: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000001',
  },
  play: async () => {
    const [single] = await screen.findAllByRole('combobox')
    expect(single).toHaveTextContent('Multi Field')

    screen.findByTestId('CheckBoxIcon')

    await screen.findByLabelText('Name')
    await screen.findByLabelText('Destination Type')
    await screen.findByLabelText('First Item')
    expect(await screen.findByPlaceholderText('11235550123')).toBeDisabled()
    await screen.findByLabelText('Second Item')
    expect(
      await screen.findByPlaceholderText('foobar@example.com'),
    ).toBeDisabled()
    await screen.findByLabelText('Third Item')
    expect(await screen.findByPlaceholderText('slack user ID')).toBeDisabled()
  },
}

export const StatusUpdates: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000000',
  },
  play: async () => {
    screen.findByTestId('CheckBoxOutlineBlankIcon')

    await waitFor(
      async () => {
        await userEvent.click(
          await screen.getByTitle(
            'Alert status updates are sent when an alert is acknowledged, closed, or escalated.',
          ),
        )
      },
      { timeout: 5000 },
    )
    await screen.findByTestId('CheckBoxIcon')
  },
}

export const ErrorField: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000000',
  },

  play: async () => {
    await userEvent.clear(await screen.findByLabelText('Name'))
    await userEvent.type(await screen.findByLabelText('Name'), 'error-test')
    await userEvent.type(
      await screen.findByPlaceholderText('11235550123'),
      '123',
    )

    const submitButton = await screen.findByText('Submit')
    await userEvent.click(submitButton)

    // response should set error on all fields plus the generic error
    await waitFor(
      async () => {
        await expect(await screen.findByLabelText('Name')).toBeInvalid()

        await expect(await screen.findByText('Name error')).toBeVisible()

        await expect(
          await screen.findByText('This indicates an invalid destination type'),
        ).toBeVisible()
        await expect(await screen.findByLabelText('Phone Number')).toBeInvalid()
        await expect(
          await screen.findByText('This is a dest field-error'),
        ).toBeVisible()

        await expect(
          await screen.findByText('This is a generic error'),
        ).toBeVisible()
      },
      { timeout: 5000 },
    )
  },
}
