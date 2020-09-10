import { round } from 'lodash-es'
import seedrandom from 'seedrandom'

type Color = [number, number, number]

// getAllyColors generates a set of n random colors with a
// contrast ratio of at least 4.5:1 against a white background
// as per WCAG standards
export function getAllyColors(seed: string, num = 1): Color[] {
  const seedRng = seedrandom(seed)

  let colors = []
  for (let i = 0; i < num; i++) {
    const colorSeed = seedrandom((seedRng() + i).toString())

    // every color needs an red, green, and blue value from 0-255
    let rgb: Color = [0, 0, 0]
    for (let j = 0; j < 3; j++) {
      const rng = seedrandom((colorSeed() + j).toString())()
      rgb[j] = Math.floor(((rng + 1) * 255) / 2)
    }

    const white: Color = [255, 255, 255]
    console.log(`color ${i + 1}`)
    console.log('%c       ', `background: rgb(${rgb[0]}, ${rgb[1]}, ${rgb[2]})`)
    console.log(`rgb: ${rgb[0]}, ${rgb[1]}, ${rgb[2]}`)
    console.log(`a11y? ${isA11y(white, rgb)}`)
    console.log(`initial contrast: ${contrastRatio(white, rgb)}:1`)

    console.log('improving contrast...')
    rgb = makeColorA11y(rgb)
    console.log('done!')

    console.log('%c       ', `background: rgb(${rgb[0]}, ${rgb[1]}, ${rgb[2]})`)
    console.log(`rgb: ${rgb[0]}, ${rgb[1]}, ${rgb[2]}`)
    console.log(`a11y? ${isA11y(white, rgb)}`)
    console.log(`final contrast: ${contrastRatio(white, rgb)}:1`)
    console.log('\n')

    colors.push(rgb)
  }

  return colors
}

// isAlly returns true/false for whether or not two relative
// luminances have a ratio higher than 4.5 as per WCAG standards
function isA11y(lightRgb: Color, darkRgb: Color): boolean {
  return contrastRatio(lightRgb, darkRgb) >= 4.5
}

// makeColorA11y takes an rgb color and adjusts the lightness in 5% increments
// or decrements until the color passes WCAG standards against a white background
function makeColorA11y(rgb: Color, adjust = 5, maxTries = 100 / adjust): Color {
  if (maxTries == 0) {
    console.log('limit reached, enjoy darkness: ', rgb)
    return [0, 0, 0]
  }

  const bgColor: Color = [255, 255, 255]

  if (isA11y(bgColor, rgb)) {
    return rgb
  }

  const curContrast = contrastRatio(bgColor, rgb)
  const _rgb = adjustBrightness(rgb, adjust)
  const nxtContrast = contrastRatio(bgColor, _rgb)

  console.log(`\t${nxtContrast}`)

  // contrast is improving if an increase is seen,
  // use previous adjust value
  if (curContrast < nxtContrast) {
    return makeColorA11y(_rgb, adjust, maxTries - 1)
  }

  // darken color instead if contrast worsens after lightening
  return makeColorA11y(_rgb, Math.abs(adjust) * -1, maxTries - 1)
}

// adjustBrightness adjusts the brightness of an rgb color
function adjustBrightness(rgb: Color, percentage: number): Color {
  let hsl = rgbToHsl(rgb)
  hsl[2] += percentage
  return hslToRgb(hsl)
}

// luminance returns the relative luminance of a color as defined by w3
// @params: rgb: [r, g, b]
// w3.org/TR/2008/REC-WCAG20-20081211/#relativeluminancedef
function luminance(rgb: Color): number {
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
function contrastRatio(lightRgb: Color, darkRgb: Color): number {
  const lightLum = luminance(lightRgb)
  const darkLum = luminance(darkRgb)

  return round((lightLum + 0.05) / (darkLum + 0.05), 1)
}

// rgbToHsl takes an rgb color and converts it to its hsl counterpart
// @params: rgb: [r, g, b]
// based off of css-tricks.com/converting-color-spaces-in-javascript/#rgb-to-hsl
function rgbToHsl(rgb: Color): Color {
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
function hslToRgb(hsl: Color): Color {
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
