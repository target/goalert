import { sortBy } from 'lodash-es'

export function formatNotificationRule(
  delayMinutes,
  { type, name, formattedValue },
) {
  const delayStr = delayMinutes
    ? `After ${delayMinutes} minute${delayMinutes === 1 ? '' : 's'}`
    : 'Immediately'

  return `${delayStr} notify me via ${type} at ${formattedValue} (${name})`
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
