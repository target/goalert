import { sortBy } from 'lodash-es'

export function formatCMValue(type, value) {
  switch (type) {
    case 'SMS':
    case 'VOICE':
      return formatPhoneNumber(value)
  }

  return value
}

// We are using libphonenumber 'International format'
// https://github.com/googlei18n/libphonenumber
// See JavaScript demo
export function formatPhoneNumber(n) {
  if (n.startsWith('+1')) {
    return `+1 ${n.slice(2, 5)}-${n.slice(5, 8)}-${n.slice(8)}`
  }
  if (n.startsWith('+91')) {
    return `+91 ${n.slice(3, 6)} ${n.slice(6, 9)} ${n.slice(9)}`
  }
  if (n.startsWith('+44')) {
    return `+44 ${n.slice(3, 7)} ${n.slice(7)}`
  }

  return n
}

export function formatNotificationRule(delayMinutes, { type, name, value }) {
  const delayStr = delayMinutes
    ? `After ${delayMinutes} minute${delayMinutes === 1 ? '' : 's'}`
    : 'Immediately'

  return `${delayStr} notify me via ${type} at ${formatCMValue(
    type,
    value,
  )} (${name})`
}

export function sortNotificationRules(nr) {
  return sortBy(nr, [
    'delayMinutes',
    'contactMethod.name',
    'contactMethod.type',
  ])
}

export function sortContactMethods(cm) {
  return sortBy(cm, ['name', 'type'])
}

export function getCountryCode(phone) {
  return ['+1', '+91', '+44'].find(cc => phone.startsWith(cc))
}

export function stripCountryCode(phone) {
  const cc = getCountryCode(phone)

  return phone.slice(cc.length)
}
