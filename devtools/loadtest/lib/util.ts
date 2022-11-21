import Chance from 'chance'

// Instantiate Chance so it can be used
var gen = new Chance()

export function genTZ(): string {
  return gen.pickone([
    'America/New_York',
    'America/Chicago',
    'America/Denver',
    'America/Los_Angeles',
    'America/Anchorage',
    'America/Adak',
    'Pacific/Honolulu',
    'Pacific/Midway',
    'Etc/UTC',
  ])
}
