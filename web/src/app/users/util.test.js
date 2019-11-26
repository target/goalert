import { formatNotificationRule } from './util'

test('formatNotificationRule', () => {
  expect(
    formatNotificationRule(0, {
      type: 'SMS',
      name: 'test',
      formattedValue: '+1 763-555-0100',
    }),
  ).toBe('Immediately notify me via SMS at +1 763-555-0100 (test)')
  expect(
    formatNotificationRule(0, {
      type: 'VOICE',
      name: 'cell',
      formattedValue: '+1 208-555-0105',
    }),
  ).toBe('Immediately notify me via VOICE at +1 208-555-0105 (cell)')
  expect(
    formatNotificationRule(10, {
      type: 'VOICE',
      name: 'phone',
      formattedValue: '+1 516-555-0184',
    }),
  ).toBe('After 10 minutes notify me via VOICE at +1 516-555-0184 (phone)')
  expect(
    formatNotificationRule(1, {
      type: 'VOICE',
      name: 'myPhone',
      formattedValue: '+91 140 000 0000',
    }),
  ).toBe('After 1 minute notify me via VOICE at +91 140 000 0000 (myPhone)')
  expect(
    formatNotificationRule(5, {
      type: 'VOICE',
      name: 'myPhone',
      formattedValue: '+44 7700 000000',
    }),
  ).toBe('After 5 minutes notify me via VOICE at +44 7700 000000 (myPhone)')
})
