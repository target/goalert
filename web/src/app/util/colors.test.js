import {
  getAllyColors, // (seed: string, num = 1): Color[]
  isA11y, // (lightRgb: Color, darkRgb: Color): boolean
  makeColorA11y, // (rgb: Color, adjust = 5, maxTries = 100 / adjust): Color
  adjustBrightness, // (rgb: Color, percentage: number): Color
  luminance, // (rgb: Color): number
  contrastRatio, // (lightRgb: Color, darkRgb: Color): number
  rgbToHsl, // (rgb: Color): Color
  hslToRgb, // (hsl: Color): Color
} from './colors'

/*
 * Using known values sourced from:
 * Contrast and luminance: contrast-ratio.com
 * HSL: rapidtables.com
 */

test('getAllyColors should generate n accessible colors from a psuedorandom source', () => {
  // generate 1 color, check with isA11y
  // generate 4 colors, check with isA11y
})

test('isA11y should determine if an rgb color meets WCAG contrast standards', () => {
  // white, black = true
  // white, white = false
  // black, black = false
  // red on white = false
  // blue on white = true
  // red on black = true
  // blue on black = false
})

test('makeColorA11y should increase the contrast ratio of of a color until a value of at least 4.5 is met', () => {
  // start with non-a11y, make a11y, check contrasts
  // start with a11y, color should not change
})

test('adjustBrightness should adjust the lightness of an rgb value by the given percentage', () => {
  // assert hsl of rgb where l = % higher than before alerting
})

test('luminance should calculate the relative luminance of an rgb value', () => {
  // white = 1
  // black = 0
  // red = 0.21
  // blue = 0.07
  // purple = 0.06
})

test('contrastRatio should return the contrast value between two rgb colors', () => {
  // white, black = 21:1
  // white, white = 1
  // black, black = 1
  // red on white = 4
  // blue on white = 8.6
  // red on black = 5.3
  // blue on black = 2.4
})

test('rgbToHsl should convert an rgb color into its known hsl value', () => {
  // white [255, 255, 255] = [0, 0, 100]
  // black [0, 0, 0] = [0, 0, 0]
  // orange [255, 165, 0] = [39, 100, 50]
  // green [0, 128, 0] = [120, 100, 25.1]
  // pink [255, 192, 203] = [350, 100, 87.6]
})

test('hslToRgb should convert an hsl color into its known rgb value', () => {
  // white [0, 0, 100] = [255, 255, 255]
  // black [0, 0, 0] = [0, 0, 0]
  // turquoise [174, 72.1, 56.5] = [64, 224, 208]
  // red [0, 100, 50] = [255, 0, 0]
  // grey [0, 0, 50.2] = [128, 128, 128]
})
