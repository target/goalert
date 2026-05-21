import { DateTime } from 'luxon'
import { Chance } from 'chance'
import { WizardFormValue } from './WizardForm'
const c = new Chance()

const keys = [
  {
    label: 'Generic API',
    value: 'generic',
  },
  {
    label: 'Grafana',
    value: 'grafana',
  },
  {
    label: 'Site24x7 Webhook URL',
    value: 'site24x7',
  },
  {
    label: 'Prometheus Alertmanager',
    value: 'prometheusAlertmanager',
  },
  {
    label: 'Email',
    value: 'email',
  },
]

const users = [
  '50322144-1e88-43dc-b638-b16a5be7bad6',
  'dfcc0684-f045-4a9f-8931-56da8a014a44',
  '016d5895-b20f-42fd-ad6c-7f1e4c11354d',
]

const timeZones = ['America/Chicago', 'Africa/Accra', 'Etc/UTC']

// scheds w/ users
export const usersSchedules: WizardFormValue = {
  teamName: 'Test',
  delayMinutes: c.integer({ min: 1, max: 9000 }),
  repeat: c.integer({ min: 1, max: 5 }).toString(),
  key: c.pickone(keys),
  primarySchedule: {
    timeZone: c.pickone(timeZones),
    users,
    rotation: {
      startDate: DateTime.local().startOf('day').toISO(),
      type: 'never',
    },
    followTheSunRotation: {
      enable: 'no',
      users: [],
      timeZone: null,
    },
  },
  secondarySchedule: {
    enable: 'yes',
    timeZone: c.pickone(timeZones),
    users,
    rotation: {
      startDate: DateTime.local().startOf('day').toISO(),
      type: 'never',
    },
    followTheSunRotation: {
      enable: 'no',
      users: [],
      timeZone: null,
    },
  },
}

// scheds w/ rotations (no fts)
export const rotationsNoFTS: WizardFormValue = {
  teamName: 'Test',
  delayMinutes: c.integer({ min: 1, max: 9000 }),
  repeat: c.integer({ min: 1, max: 5 }).toString(),
  key: c.pickone(keys),
  primarySchedule: {
    timeZone: 'America/Chicago',
    users,
    rotation: {
      startDate: DateTime.local().startOf('day').toISO(),
      type: 'daily',
    },
    followTheSunRotation: {
      enable: 'no',
      users: [],
      timeZone: null,
    },
  },
  secondarySchedule: {
    enable: 'yes',
    timeZone: 'Africa/Accra',
    users,
    rotation: {
      startDate: DateTime.local().startOf('day').toISO(),
      type: 'weekly',
    },
    followTheSunRotation: {
      enable: 'no',
      users: [],
      timeZone: null,
    },
  },
}

// scheds w/ rotations + fts
export const rotationsAndFTS: WizardFormValue = {
  teamName: 'Test',
  delayMinutes: c.integer({ min: 1, max: 9000 }),
  repeat: c.integer({ min: 1, max: 5 }).toString(),
  key: c.pickone(keys),
  primarySchedule: {
    timeZone: 'Etc/UTC',
    users,
    rotation: {
      startDate: DateTime.local().startOf('day').toISO(),
      type: 'weekly',
    },
    followTheSunRotation: {
      enable: 'yes',
      users,
      timeZone: 'America/Chicago',
    },
  },
  secondarySchedule: {
    enable: 'yes',
    timeZone: 'Africa/Accra',
    users,
    rotation: {
      startDate: DateTime.local().startOf('day').toISO(),
      type: 'daily',
    },
    followTheSunRotation: {
      enable: 'yes',
      users,
      timeZone: 'Africa/Accra',
    },
  },
}
