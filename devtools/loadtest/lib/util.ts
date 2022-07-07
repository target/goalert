import Chance from 'chance'

// Instantiate Chance so it can be used
var gen = new Chance()

export function genTZ(): string {
  const genTZ = gen.timezone()
  return genTZ.utc ? genTZ.utc[0] : 'Etc/UTC'
}
