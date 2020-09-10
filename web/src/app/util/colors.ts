import { round } from 'lodash-es'
import seedrandom from 'seedrandom'

type Colors = { r: number; g: number; b: number }[]

// getColors generates a set of n random colors based off
// of a seeded random number generator from the string.
// a random int32 value is used to linearly map the rgb values
export function getColors(seed: string, num = 1): Colors {
  const seedRng = seedrandom(seed)

  let colors = []
  for (let i = 0; i < num; i++) {
    const colorSeed = seedrandom((seedRng() + i).toString())

    // every color needs an red, green, and blue value from 0-255
    let rgb = [] // [r, g, b]
    for (let j = 0; j < 3; j++) {
      const rng = seedrandom((colorSeed() + j).toString())()
      rgb[j] = Math.floor(((rng + 1) * 255) / 2)
    }

    const whiteLum = luminance([255, 255, 255]) // dialog bg color
    const lum = luminance(rgb)
    const contrast = round(contrastRatio(whiteLum, lum), 1)
    const isAlly = isA11y(whiteLum, lum)

    console.log(`color ${i + 1}`)
    console.log('%c       ', `background: rgb(${rgb[0]}, ${rgb[1]}, ${rgb[2]})`)
    console.log(`rgb: ${rgb[0]}, ${rgb[1]}, ${rgb[2]}`)
    console.log(`contrast ratio: ${contrast}`)
    console.log(`a11y? ${isAlly}`)
    console.log('\n')

    colors.push({
      r: rgb[0],
      g: rgb[1],
      b: rgb[2],
    })
  }

  return colors
}

// luminance returns the relative luminance of a color as defined by w3
// @params: rgb: [r, g, b]
// w3.org/TR/2008/REC-WCAG20-20081211/#relativeluminancedef
function luminance(rgb: number[]): number {
  // calculate sRGB values
  const srgb = rgb.map((x) => x / 255)

  // calculate R, G, and B values for rel luminance equation
  const RGB = srgb.map((x) =>
    x <= 0.03928 ? x / 12.92 : ((x + 0.055) / 1.055) ** 2.4,
  )

  // calculate relative luminance
  return 0.2126 * RGB[0] + 0.7152 * RGB[1] + 0.0722 * RGB[2]
}

// contrastRatio returns the contrast ratio x:1 as defined by w3
// w3.org/TR/2008/REC-WCAG20-20081211/#contrast-ratiodef
function contrastRatio(lightLum: number, darkLum: number): number {
  return (lightLum + 0.05) / (darkLum + 0.05)
}

// isAlly returns true/false for whether or not two
// relative luminances have a ratio higher than 4.5
// as per WCAG standards
function isA11y(lightLum: number, darkLum: number): boolean {
  return contrastRatio(lightLum, darkLum) >= 4.5
}

// rgbToHsl takes an rgb color and converts it to its hsl counterpart
// @params: rgb: [r, g, b]
// based off of css-tricks.com/converting-color-spaces-in-javascript/#rgb-to-hsl
function rgbToHsl(rgb: number[]): number[] {
  // calculate sRGB values
  const srgb = rgb.map((x) => x / 255)
  const r = srgb[0]
  const g = srgb[1]
  const b = srgb[2]

  // find greatest and smallest channel values
  const cmin = Math.min(r, g, b)
  const cmax = Math.max(r, g, b)
  const delta = cmax - cmin
  let h = 0
  let s = 0
  let l = 0

  // calculate hue
  if (delta == 0) h = 0
  else if (cmax == r) h = ((g - b) / delta) % 6
  else if (cmax == g) h = (b - r) / delta + 2
  else h = (r - g) / delta + 4
  h = Math.round(h * 60)

  // make negative hues positive behind 360°
  if (h < 0) h += 360

  l = (cmax + cmin) / 2 // calculate lightness
  s = delta == 0 ? 0 : delta / (1 - Math.abs(2 * l - 1)) // calculate saturation

  // calculate saturation and lightness as percentages
  s = round(s * 100, 1)
  l = round(l * 100, 1)

  return [h, s, l]
}

// rgbToHsl takes an hsl color and converts it to its rgb counterpart
// @params: hsl: [h, s, l]
// based off of css-tricks.com/converting-color-spaces-in-javascript/#hsl-to-rgb
function hslToRgb(hsl: number[]): number[] {
  const h = hsl[0]
  let s = hsl[1]
  let l = hsl[2]

  // convert from percentage
  s /= 100
  l /= 100

  let c = (1 - Math.abs(2 * l - 1)) * s,
    x = c * (1 - Math.abs(((h / 60) % 2) - 1)),
    m = l - c / 2,
    r = 0,
    g = 0,
    b = 0

  if (0 <= h && h < 60) {
    r = c
    g = x
    b = 0
  } else if (60 <= h && h < 120) {
    r = x
    g = c
    b = 0
  } else if (120 <= h && h < 180) {
    r = 0
    g = c
    b = x
  } else if (180 <= h && h < 240) {
    r = 0
    g = x
    b = c
  } else if (240 <= h && h < 300) {
    r = x
    g = 0
    b = c
  } else if (300 <= h && h < 360) {
    r = c
    g = 0
    b = x
  }

  r = round((r + m) * 255)
  g = round((g + m) * 255)
  b = round((b + m) * 255)

  return [r, g, b]
}
