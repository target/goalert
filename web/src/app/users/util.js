import { sortBy } from 'lodash-es'
import seedrandom from 'seedrandom'

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

// getVerifyCodeColors generates a set of four random colors based off
// of a seeded random number generator from the given verification code.
// the random int32 value is linearly mapped to rgb values
export function getVerifyCodeColors(code) {
  const codeSeed = seedrandom(code)

  let colors = []
  for (let i = 0; i < 4; i++) {
    const colorSeed = seedrandom(codeSeed() + i)
    let rgb = [] // [r, g, b]
    for (let j = 0; j < 3; j++) {
      const rng = seedrandom(colorSeed() + j)()
      rgb[j] = Math.floor(((rng + 1) * 255) / 2)
    }

    colors.push({
      r: rgb[0],
      g: rgb[1],
      b: rgb[2],
    })
  }

  return colors
}

// luminance returns the relative luminance of a color
// rgb: [r, g, b]
function luminance(rgb) {
  return 0.2126 * r + 0.7152 * b + 0.0722 * b
}

// isAlly returns true/false for whether or not two
// relative luminances have a ratio higher than 4.5
// as per WCAG standards
function isAlly(lightLum, darkLum) {}

//console.log('%c   ', `background: rgb(${rgb[0]}, ${rgb[1]}, ${rgb[2]})`)
