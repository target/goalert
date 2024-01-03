import React from 'react'
import { test, expect } from '@playwright/experimental-ct-react'
import { GQLMock } from '../../playwright/gqlMock'
import { DestinationFieldValueWrapper } from './DestinationField.story'

test.use({ viewport: { width: 500, height: 500 } })

test('should render', async ({ mount, page }) => {
  const gql = new GQLMock(page)
  await gql.init()

  gql.setGQL('DestTypes', (vars) => {
    return {
      data: {
        destinationTypes: [
          {
            type: 'builtin-twilio-sms',
            name: 'Text Message (SMS)',
            enabled: false,
            disabledMessage: 'Twilio must be configured by an administrator',
            userDisclaimer: '',
            isContactMethod: true,
            isEPTarget: false,
            isSchedOnCallNotify: false,
            requiredFields: [
              {
                fieldID: 'phone-number',
                labelSingular: 'Phone Number',
                labelPlural: 'Phone Numbers',
                hint: 'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
                hintURL: '',
                placeholderText: '11235550123',
                prefix: '+',
                inputType: 'tel',
                isSearchSelectable: false,
                supportsValidation: true,
                __typename: 'DestinationFieldConfig',
              },
            ],
            __typename: 'DestinationTypeInfo',
          },
          {
            type: 'builtin-twilio-voice',
            name: 'Voice Call',
            enabled: false,
            disabledMessage: 'Twilio must be configured by an administrator',
            userDisclaimer: '',
            isContactMethod: true,
            isEPTarget: false,
            isSchedOnCallNotify: false,
            requiredFields: [
              {
                fieldID: 'phone-number',
                labelSingular: 'Phone Number',
                labelPlural: 'Phone Numbers',
                hint: 'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
                hintURL: '',
                placeholderText: '11235550123',
                prefix: '+',
                inputType: 'tel',
                isSearchSelectable: false,
                supportsValidation: true,
                __typename: 'DestinationFieldConfig',
              },
            ],
            __typename: 'DestinationTypeInfo',
          },
          {
            type: 'builtin-smtp-email',
            name: 'Email',
            enabled: false,
            disabledMessage: 'SMTP must be configured by an administrator',
            userDisclaimer: '',
            isContactMethod: true,
            isEPTarget: false,
            isSchedOnCallNotify: false,
            requiredFields: [
              {
                fieldID: 'email-address',
                labelSingular: 'Email Address',
                labelPlural: 'Email Addresses',
                hint: '',
                hintURL: '',
                placeholderText: 'foobar@example.com',
                prefix: '',
                inputType: 'email',
                isSearchSelectable: false,
                supportsValidation: true,
                __typename: 'DestinationFieldConfig',
              },
            ],
            __typename: 'DestinationTypeInfo',
          },
          {
            type: 'builtin-webhook',
            name: 'Webhook',
            enabled: false,
            disabledMessage: 'Webhooks must be enabled by an administrator',
            userDisclaimer: '',
            isContactMethod: true,
            isEPTarget: true,
            isSchedOnCallNotify: true,
            requiredFields: [
              {
                fieldID: 'webhook-url',
                labelSingular: 'Webhook URL',
                labelPlural: 'Webhook URLs',
                hint: 'Webhook Documentation',
                hintURL: '/docs#webhooks',
                placeholderText: 'https://example.com',
                prefix: '',
                inputType: 'url',
                isSearchSelectable: false,
                supportsValidation: true,
                __typename: 'DestinationFieldConfig',
              },
            ],
            __typename: 'DestinationTypeInfo',
          },
          {
            type: 'builtin-slack-dm',
            name: 'Slack Message (DM)',
            enabled: false,
            disabledMessage: 'Slack must be enabled by an administrator',
            userDisclaimer: '',
            isContactMethod: true,
            isEPTarget: false,
            isSchedOnCallNotify: false,
            requiredFields: [
              {
                fieldID: 'slack-user-id',
                labelSingular: 'Slack User',
                labelPlural: 'Slack Users',
                hint: 'Go to your Slack profile, click the three dots, and select "Copy member ID".',
                hintURL: '',
                placeholderText: 'member ID',
                prefix: '',
                inputType: 'text',
                isSearchSelectable: false,
                supportsValidation: false,
                __typename: 'DestinationFieldConfig',
              },
            ],
            __typename: 'DestinationTypeInfo',
          },
          {
            type: 'builtin-slack-channel',
            name: 'Slack Channel',
            enabled: false,
            disabledMessage: 'Slack must be enabled by an administrator',
            userDisclaimer: '',
            isContactMethod: false,
            isEPTarget: true,
            isSchedOnCallNotify: true,
            requiredFields: [
              {
                fieldID: 'slack-channel-id',
                labelSingular: 'Slack Channel',
                labelPlural: 'Slack Channels',
                hint: '',
                hintURL: '',
                placeholderText: '',
                prefix: '',
                inputType: 'text',
                isSearchSelectable: true,
                supportsValidation: false,
                __typename: 'DestinationFieldConfig',
              },
            ],
            __typename: 'DestinationTypeInfo',
          },
          {
            type: 'builtin-slack-usergroup',
            name: 'Update Slack User Group',
            enabled: false,
            disabledMessage: 'Slack must be enabled by an administrator',
            userDisclaimer: '',
            isContactMethod: false,
            isEPTarget: false,
            isSchedOnCallNotify: true,
            requiredFields: [
              {
                fieldID: 'slack-usergroup-id',
                labelSingular: 'User Group',
                labelPlural: 'User Groups',
                hint: "The selected group's membership will be replaced/set to the schedule's on-call user(s).",
                hintURL: '',
                placeholderText: '',
                prefix: '',
                inputType: 'text',
                isSearchSelectable: true,
                supportsValidation: false,
                __typename: 'DestinationFieldConfig',
              },
              {
                fieldID: 'slack-channel-id',
                labelSingular: 'Slack Channel (for errors)',
                labelPlural: 'Slack Channels (for errors)',
                hint: 'If the user group update fails, an error will be posted to this channel.',
                hintURL: '',
                placeholderText: '',
                prefix: '',
                inputType: 'text',
                isSearchSelectable: true,
                supportsValidation: false,
                __typename: 'DestinationFieldConfig',
              },
            ],
            __typename: 'DestinationTypeInfo',
          },
        ],
      },
    }
  })

  const component = await mount(
    <DestinationFieldValueWrapper
      value={[{ fieldID: 'webhook-url', value: ' ' }]}
      destType='builtin-webhook'
    />,
  )
  // ensure text loads correctly
  await expect(component).toContainText('Webhook URL')
  await expect(component.locator('a')).toContainText('Webhook Documentation')
  await expect(component.locator('a')).toHaveAttribute('href')

  //   // ensure close icon visible for invalid urls
  //   await component.locator('input').fill('example')
  //   await expect(component.locator('input')).toHaveValue('example')
  //   await expect(component.locator('svg[data-testid="CloseIcon"]')).toBeVisible()

  //   // ensure check icon visible for valid urls
  //   await component.locator('input').fill('https://example.com')
  //   await expect(component.locator('input')).toHaveValue('https://example.com')
  //   await expect(component.locator('svg[data-testid="CheckIcon"]')).toBeVisible()
})
