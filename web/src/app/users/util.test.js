import {
  formatPhoneNumber,
  formatNotificationRule,
  getCountryCode,
  stripCountryCode,
} from './util'

test('formatPhoneNumber', () => {
  expect(formatPhoneNumber('+17635550100')).toBe('+1 763-555-0100')
  expect(formatPhoneNumber('+12085550105')).toBe('+1 208-555-0105')
  expect(formatPhoneNumber('+15165550184')).toBe('+1 516-555-0184')
  expect(formatPhoneNumber('+911400000000')).toBe('+91 14 0000 0000')
  expect(formatPhoneNumber('+911401234567')).toBe('+91 14 0123 4567')
  expect(formatPhoneNumber('+911409876543')).toBe('+91 14 0987 6543')
  expect(formatPhoneNumber('+447700000000')).toBe('+44 7700 000000')
  expect(formatPhoneNumber('+447701234567')).toBe('+44 7701 234567')
  expect(formatPhoneNumber('+447709876543')).toBe('+44 7709 876543')
})

test('formatNotificationRule', () => {
  expect(
    formatNotificationRule(0, {
      type: 'SMS',
      name: 'test',
      value: '+17635550100',
    }),
  ).toBe('Immediately notify me via SMS at +1 763-555-0100 (test)')
  expect(
    formatNotificationRule(0, {
      type: 'VOICE',
      name: 'cell',
      value: '+12085550105',
    }),
  ).toBe('Immediately notify me via VOICE at +1 208-555-0105 (cell)')
  expect(
    formatNotificationRule(10, {
      type: 'VOICE',
      name: 'phone',
      value: '+15165550184',
    }),
  ).toBe('After 10 minutes notify me via VOICE at +1 516-555-0184 (phone)')
  expect(
    formatNotificationRule(1, {
      type: 'VOICE',
      name: 'myPhone',
      value: '+911400000000',
    }),
  ).toBe('After 1 minute notify me via VOICE at +91 14 0000 0000 (myPhone)')
  expect(
    formatNotificationRule(5, {
      type: 'VOICE',
      name: 'myPhone',
      value: '+447700000000',
    }),
  ).toBe('After 5 minutes notify me via VOICE at +44 7700 000000 (myPhone)')
})

test('getCountryCode', () => {
  expect(getCountryCode('+17635550100')).toBe('+1')
  expect(getCountryCode('+12085550105')).toBe('+1')
  expect(getCountryCode('+15165550184')).toBe('+1')
  expect(getCountryCode('+911400000000')).toBe('+91')
  expect(getCountryCode('+911401234567')).toBe('+91')
  expect(getCountryCode('+911409876543')).toBe('+91')
  expect(getCountryCode('+447700000000')).toBe('+44')
  expect(getCountryCode('+447701234567')).toBe('+44')
  expect(getCountryCode('+447709876543')).toBe('+44')
})

test('stripCountryCode', () => {
  expect(stripCountryCode('+17635550100')).toBe('7635550100')
  expect(stripCountryCode('+12085550105')).toBe('2085550105')
  expect(stripCountryCode('+15165550184')).toBe('5165550184')
  expect(stripCountryCode('+911400000000')).toBe('1400000000')
  expect(stripCountryCode('+911401234567')).toBe('1401234567')
  expect(stripCountryCode('+911409876543')).toBe('1409876543')
  expect(stripCountryCode('+447700000000')).toBe('7700000000')
  expect(stripCountryCode('+447701234567')).toBe('7701234567')
  expect(stripCountryCode('+447709876543')).toBe('7709876543')
})
