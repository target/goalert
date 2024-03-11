import { DestinationTypeInfo } from '../../schema'

export const destTypes: DestinationTypeInfo[] = [
  {
    type: 'single-field',
    name: 'Single Field',
    enabled: true,
    disabledMessage: 'Single field destination type must be configured.',
    userDisclaimer: '',
    isContactMethod: true,
    isEPTarget: false,
    isSchedOnCallNotify: true,
    iconURL: '',
    iconAltText: '',
    supportsStatusUpdates: false,
    statusUpdatesRequired: false,
    requiredFields: [
      {
        fieldID: 'phone-number',
        label: 'Phone Number',
        hint: 'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
        hintURL: '',
        placeholderText: '11235550123',
        prefix: '+',
        inputType: 'tel',
        supportsSearch: false,
        supportsValidation: true,
      },
    ],
  },
  {
    type: 'triple-field',
    name: 'Multi Field',
    enabled: true,
    disabledMessage: 'Multi field destination type must be configured.',
    userDisclaimer: '',
    isContactMethod: true,
    isEPTarget: false,
    isSchedOnCallNotify: true,
    iconURL: '',
    iconAltText: '',
    supportsStatusUpdates: true,
    statusUpdatesRequired: false,
    requiredFields: [
      {
        fieldID: 'first-field',
        label: 'First Item',
        hint: 'Some hint text',
        hintURL: '',
        placeholderText: '11235550123',
        prefix: '+',
        inputType: 'tel',
        supportsSearch: false,
        supportsValidation: true,
      },
      {
        fieldID: 'second-field',
        label: 'Second Item',
        hint: '',
        hintURL: '',
        placeholderText: 'foobar@example.com',
        prefix: '',
        inputType: 'email',
        supportsSearch: false,
        supportsValidation: true,
      },
      {
        fieldID: 'third-field',
        label: 'Third Item',
        hint: 'docs',
        hintURL: '/docs',
        placeholderText: 'slack user ID',
        prefix: '',
        inputType: 'string',
        supportsSearch: false,
        supportsValidation: true,
      },
    ],
  },
  {
    type: 'supports-status',
    name: 'Single With Status',
    enabled: true,
    disabledMessage: 'Single field destination type must be configured.',
    userDisclaimer: '',
    isContactMethod: true,
    isEPTarget: false,
    isSchedOnCallNotify: false,
    iconURL: '',
    iconAltText: '',
    supportsStatusUpdates: true,
    statusUpdatesRequired: false,
    requiredFields: [
      {
        fieldID: 'phone-number',
        label: 'Phone Number',
        hint: 'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
        hintURL: '',
        placeholderText: '11235550123',
        prefix: '+',
        inputType: 'tel',
        supportsSearch: false,
        supportsValidation: true,
      },
    ],
  },
  {
    type: 'required-status',
    name: 'Single With Required Status',
    enabled: true,
    disabledMessage: 'Single field destination type must be configured.',
    userDisclaimer: '',
    isContactMethod: true,
    isEPTarget: false,
    isSchedOnCallNotify: false,
    iconURL: '',
    iconAltText: '',
    supportsStatusUpdates: false,
    statusUpdatesRequired: true,
    requiredFields: [
      {
        fieldID: 'phone-number',
        label: 'Phone Number',
        hint: 'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
        hintURL: '',
        placeholderText: '11235550123',
        prefix: '+',
        inputType: 'tel',
        supportsSearch: false,
        supportsValidation: true,
      },
    ],
  },
  {
    type: 'multi-field-ep-step',
    name: 'Multi Field EP Step Dest',
    enabled: true,
    disabledMessage: 'Multi field EP destination type must be configured.',
    userDisclaimer: '',
    isContactMethod: false,
    isEPTarget: true,
    isSchedOnCallNotify: false,
    iconURL: '',
    iconAltText: '',
    supportsStatusUpdates: false,
    statusUpdatesRequired: true,
    requiredFields: [
      {
        fieldID: 'phone-number',
        label: 'Phone Number',
        hint: 'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
        hintURL: '',
        placeholderText: '11235550123',
        prefix: '+',
        inputType: 'tel',
        supportsSearch: false,
        supportsValidation: true,
      },
      {
        fieldID: 'webhook-url',
        label: 'Webhook URL',
        hint: 'Webhook Documentation',
        hintURL: '/docs#webhooks',
        placeholderText: 'https://example.com',
        prefix: '',
        inputType: 'url',
        supportsSearch: false,
        supportsValidation: true,
      },
    ],
  },
  {
    type: 'dest-type-error-ep-step',
    name: 'Dest Type Error EP Step',
    enabled: true,
    disabledMessage: '',
    userDisclaimer: '',
    isContactMethod: false,
    isEPTarget: true,
    isSchedOnCallNotify: false,
    iconURL: '',
    iconAltText: '',
    supportsStatusUpdates: false,
    statusUpdatesRequired: true,
    requiredFields: [
      {
        fieldID: 'phone-number',
        label: 'Phone Number',
        hint: 'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
        hintURL: '',
        placeholderText: '11235550123',
        prefix: '+',
        inputType: 'tel',
        supportsSearch: false,
        supportsValidation: true,
      },
    ],
  },
  {
    type: 'generic-error-ep-step',
    name: 'Generic Error EP Step',
    enabled: true,
    disabledMessage: '',
    userDisclaimer: '',
    isContactMethod: false,
    isEPTarget: true,
    isSchedOnCallNotify: false,
    iconURL: '',
    iconAltText: '',
    supportsStatusUpdates: false,
    statusUpdatesRequired: true,
    requiredFields: [
      {
        fieldID: 'phone-number',
        label: 'Phone Number',
        hint: 'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
        hintURL: '',
        placeholderText: '11235550123',
        prefix: '+',
        inputType: 'tel',
        supportsSearch: false,
        supportsValidation: true,
      },
    ],
  },
  {
    type: 'disabled-destination',
    name: 'This is disabled',
    enabled: false,
    disabledMessage: 'This field is disabled.',
    userDisclaimer: '',
    isContactMethod: true,
    isEPTarget: true,
    isSchedOnCallNotify: true,
    iconURL: '',
    iconAltText: '',
    supportsStatusUpdates: false,
    statusUpdatesRequired: false,
    requiredFields: [
      {
        fieldID: 'disabled',
        label: '',
        hint: '',
        hintURL: '',
        placeholderText: 'This field is disabled.',
        prefix: '',
        inputType: 'url',
        supportsSearch: false,
        supportsValidation: false,
      },
    ],
  },
]
