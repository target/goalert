import p from 'prop-types'

const scheduleShape = p.shape({
  timeZone: p.string,
  users: p.array,
  rotation: p.shape({
    startDate: p.object, // DateTime, null when empty val
    type: p.string,
  }),
  followTheSunRotation: p.shape({
    enable: p.string,
    users: p.array,
    timeZone: p.string,
  }),
})

export const value = p.shape({
  teamName: p.string,
  primarySchedule: scheduleShape,
  secondarySchedule: scheduleShape,
  delayMinutes: p.string,
  repeat: p.string,
  key: p.shape({
    label: p.string,
    value: p.string,
  }),
}).isRequired
