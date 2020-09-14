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

test('getAllyColors should generate n accessible colors from a psuedorandom source', () => {})
test('isA11y should determine if an rgb color meets WCAG contrast standards', () => {})
test('makeColorA11y should increase the contrast ratio of of a color until a value of at least 4.5 is met', () => {})
test('adjustBrightness should adjust the lightness of an rgb value by the given percentage', () => {})
test('luminance should calculate the relative luminance of an rgb value', () => {})

test('contrastRatio should return the contrast value between two rgb colors', () => {
  // white, black
  // white, white
  // black, black
  // color, white
  // color, black
})
test('rgbToHsl should convert an rgb color into its known hsl value', () => {})
test('hslToRgb should convert an hsl color into its known rgb value', () => {})
