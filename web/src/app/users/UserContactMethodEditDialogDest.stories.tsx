import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodEditDialogDest from './UserContactMethodEditDialogDest'
import { expect, fn, userEvent, waitFor, within } from '@storybook/test'
import { useArgs } from '@storybook/preview-api'
import { DestFieldValueError, InputFieldError } from '../util/errtypes'
import {
  Destination,
  DestinationFieldValidateInput,
  UpdateUserContactMethodInput,
} from '../../schema'
import { mockOp } from '../storybook/graphql'

const meta = {
  title: 'users/UserContactMethodEditDialogDest',
  component: UserContactMethodEditDialogDest,
  tags: ['autodocs'],
  args: {
    onClose: fn(),
  },
  parameters: {
    docs: {
      story: {
        inline: false,
        iframeHeight: 500,
      },
    },
    fetchMock: {
      mocks: [
        mockOp<unknown, { id: string }>('userCm', (vars) => {
          return {
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
                          },
                        ],
                        displayInfo: {
                          text: '+1 555-555-5555',
                          iconAltText: 'Voice Call',
                          iconURL: '',
                          linkURL: '',
                        },
                      } satisfies Destination,
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
                            value: '+11235550123',
                          },
                          {
                            fieldID: 'second-field',
                            value: 'foobar@example.com',
                          },
                          {
                            fieldID: 'third-field',
                            value: 'slack',
                          },
                        ],
                        displayInfo: {
                          text: '11235550123',
                          iconAltText: 'Mulitple Fields Example',
                          iconURL: '',
                          linkURL: '',
                        },
                      } satisfies Destination,
                      statusUpdates: 'ENABLED',
                      disabled: false,
                      pending: false,
                    },
            },
          }
        }),
        mockOp<UpdateUserContactMethodInput>(
          'UpdateUserContactMethod',
          (vars) => {
            if (vars.input.name === 'error-test') {
              return {
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
              }
            }
            return {
              data: {
                updateUserContactMethod: {
                  id: '00000000-0000-0000-0000-000000000000',
                },
              },
            }
          },
        ),
        mockOp<DestinationFieldValidateInput>('ValidateDestination', (vars) => {
          return {
            data: {
              destinationFieldValidate:
                vars.input.value === '@slack' ||
                vars.input.value === '+12225558989' ||
                vars.input.value === 'valid@email.com',
            },
          }
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
    return (
      <UserContactMethodEditDialogDest
        {...args}
        disablePortal
        onClose={onClose}
      />
    )
  },
} satisfies Meta<typeof UserContactMethodEditDialogDest>

export default meta
type Story = StoryObj<typeof meta>

export const SingleField: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000000',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.click(await canvas.findByLabelText('Destination Type'))

    const single = await canvas.findByLabelText('Destination Type')
    expect(single).toHaveTextContent('Single With Status')
    await canvas.findByTestId('CheckBoxOutlineBlankIcon')
  },
}

export const MultiField: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000001',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const multi = await canvas.findByLabelText('Destination Type')
    expect(multi).toHaveTextContent('Multi Field')

    canvas.findByTestId('CheckBoxIcon')

    await canvas.findByLabelText('Name')
    await canvas.findByLabelText('Destination Type')
    await canvas.findByLabelText('First Item')
    expect(await canvas.findByPlaceholderText('11235550123')).toBeDisabled()
    await canvas.findByLabelText('Second Item')
    expect(
      await canvas.findByPlaceholderText('foobar@example.com'),
    ).toBeDisabled()
    await canvas.findByLabelText('Third Item')
    expect(await canvas.findByPlaceholderText('slack user ID')).toBeDisabled()
  },
}

export const StatusUpdates: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000000',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    canvas.findByTestId('CheckBoxOutlineBlankIcon')

    await waitFor(
      async () => {
        await userEvent.click(
          await canvas.getByTitle(
            'Alert status updates are sent when an alert is acknowledged, closed, or escalated.',
          ),
        )
      },
      { timeout: 5000 },
    )
    await canvas.findByTestId('CheckBoxIcon')
  },
}

export const ErrorField: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000000',
  },

  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.clear(await canvas.findByLabelText('Name'))
    await userEvent.type(await canvas.findByLabelText('Name'), 'error-test')
    await userEvent.type(
      await canvas.findByPlaceholderText('11235550123'),
      '123',
    )

    const submitButton = await canvas.findByText('Submit')
    await userEvent.click(submitButton)

    // response should set error on all fields plus the generic error
    await waitFor(
      async () => {
        await expect(await canvas.findByLabelText('Name'))

        await expect(await canvas.findByText('Name error')).toBeVisible()

        await expect(
          await canvas.findByText('This indicates an invalid destination type'),
        ).toBeVisible()
        await expect(await canvas.findByLabelText('Phone Number'))
        await expect(
          await canvas.findByText('This is a dest field-error'),
        ).toBeVisible()

        await expect(
          await canvas.findByText('This is a generic error'),
        ).toBeVisible()
      },
      { timeout: 5000 },
    )
  },
}
