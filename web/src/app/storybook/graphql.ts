import { GraphQLHandler, HttpResponse, graphql } from 'msw'
import {
  ConfigID,
  ConfigType,
  IntegrationKeyTypeInfo,
  UserRole,
} from '../../schema'

export type ConfigItem = {
  id: ConfigID
  type: ConfigType
  value: string
}

export type RequireConfigDoc = {
  user: {
    id: string
    name: string
    role: UserRole
  }
  config: ConfigItem[]
  integrationKeyTypes: IntegrationKeyTypeInfo[]
}

export function handleConfig(doc: RequireConfigDoc): GraphQLHandler {
  return graphql.query('RequireConfig', () => {
    return HttpResponse.json({
      data: doc,
    })
  })
}

export const defaultConfig: RequireConfigDoc = {
  user: {
    id: '00000000-0000-0000-0000-000000000000',
    name: 'Test User',
    role: 'admin',
  },
  config: [
    {
      id: 'General.ApplicationName',
      type: 'string',
      value: 'ComponentTest',
    },
  ],
  integrationKeyTypes: [
    {
      id: 'example-enabled',
      name: 'Enabled Example',
      label: 'Enabled Example Value',
      enabled: true,
    },
    {
      id: 'example-disabled',
      name: 'Disabled Example',
      label: 'Disabled Example Value',
      enabled: true,
    },
  ],
}

export const handleDefaultConfig = handleConfig(defaultConfig)

export function handleDestTypes(doc: DestinationTypeInfo[]): GraphQLHandler {
  return graphql.query('DestTypes', () => {
    return HttpResponse.json({
      data: {
        destinationTypes: doc,
      },
    })
  })
}

export const destTypes: DestinationTypeInfo[] = [
  {
    type: 'builtin-twilio-sms',
    name: 'Text Message (SMS)',
    enabled: true,
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
    enabled: true,
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
    enabled: true,
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
    enabled: true,
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
    enabled: true,
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
    enabled: true,
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
    enabled: true,
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
]

export const handleDefaultDestTypes = handleDestTypes(destTypes)
