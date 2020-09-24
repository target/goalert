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
  // generate 1 color
  const generateOneColor = getAllyColors(`zxyjhkjh`, 1)
  expect(generateOneColor).toHaveLength(1)
  // generate 4 colors
  const generateFourColors = getAllyColors(`kljahsdf`, 4)
  expect(generateFourColors).toHaveLength(4)
})

test('isA11y should determine if an rgb color meets WCAG contrast standards', () => {
  // white, black = true
  const whiteBlackContrast = isA11y([255, 255, 255], [0, 0, 0])
  expect(whiteBlackContrast).toBe(true)
  // white, white = false
  const whiteWhiteContrast = isA11y([255, 255, 255], [255, 255, 255])
  expect(whiteWhiteContrast).toBe(false)
  // black, black = false
  const blackBlackContrast = isA11y([0, 0, 0], [0, 0, 0])
  expect(blackBlackContrast).toBe(false)
  // red on white = false
  const redOnWhiteContrast = isA11y([255, 0, 0], [255, 255, 255])
  expect(redOnWhiteContrast).toBe(false)
  // blue on white = true
  const blueOnWhiteContrast = isA11y([255, 255, 255], [0, 0, 255])
  expect(blueOnWhiteContrast).toBe(true)
  // red on black = true
  const redOnBlackContrast = isA11y([255, 0, 0], [0, 0, 0])
  expect(redOnBlackContrast).toBe(true)
  // blue on black = false
  const blueOnBlackContrast = isA11y([0, 0, 255], [0, 0, 0])
  expect(blueOnBlackContrast).toBe(false)
})

test('makeColorA11y should increase the contrast ratio of of a color until a value of at least 4.5 is met', () => {
  // start with non-a11y, make a11y, check contrasts
  const nonA11yContrastRatio = makeColorA11y([201, 201, 201], 30, 6)
  const checkNonA11yContrast = isA11y([201, 201, 201], nonA11yContrastRatio)
  expect(checkNonA11yContrast).toBe(true)
  // start with a11y, color should not change
  const a11yContrastRatio = makeColorA11y([128, 0, 128], 5, 10)
  expect(a11yContrastRatio).toEqual([128, 0, 128])
})

test('adjustBrightness should adjust the lightness of an rgb value by the given percentage', () => {
  // assert hsl of rgb where l = % higher than before alerting
  const adjustPurpleBrightness = adjustBrightness([128, 0, 128], 10)
  expect(adjustPurpleBrightness).toEqual([179, 0, 179])
})

test('luminance should calculate the relative luminance of an rgb value', () => {
  // white = 1
  const whiteLuminance = luminance([255, 255, 255])
  expect(whiteLuminance).toBe(1)
  // black = 0
  const blackLuminance = luminance([0, 0, 0])
  expect(blackLuminance).toBe(0)
  // red = 0.21
  const redLuminance = luminance([255, 0, 0])
  expect(redLuminance).toBe(0.2126)
  // blue = 0.07
  const blueLuminance = luminance([0, 0, 255])
  expect(blueLuminance).toBe(0.0722)
  // purple = 0.06
  const purpleLuminance = luminance([128, 0, 128])
  expect(purpleLuminance).toBe(0.06147707043243851)
})

test('contrastRatio should return the contrast value between two rgb colors', () => {
  // white, black = 21:1
  const whiteOnBlack = contrastRatio([255, 255, 255], [0, 0, 0])
  expect(whiteOnBlack).toBe(21)
  // white, white = 1
  const whiteOnWhite = contrastRatio([255, 255, 255], [255, 255, 255])
  expect(whiteOnWhite).toBe(1)
  // black, black = 1
  const blackOnBlack = contrastRatio([0, 0, 0], [0, 0, 0])
  expect(blackOnBlack).toBe(1)
  // red on white = 4
  const redOnWhite = contrastRatio([255, 255, 255], [255, 0, 0])
  expect(redOnWhite).toBe(4)
  // blue on white = 8.6
  const blueOnWhite = contrastRatio([255, 255, 255], [0, 0, 255])
  expect(blueOnWhite).toBe(8.6)
  // red on black = 5.3
  const redOnBlack = contrastRatio([255, 0, 0], [0, 0, 0])
  expect(redOnBlack).toBe(5.3)
  // blue on black = 2.4
  const blueOnBlack = contrastRatio([0, 0, 255], [0, 0, 0])
  expect(blueOnBlack).toBe(2.4)
})

test('rgbToHsl should convert an rgb color into its known hsl value', () => {
  // white [255, 255, 255] = [0, 0, 100]
  const convertWhiteToHsl = rgbToHsl([255, 255, 255])
  expect(convertWhiteToHsl).toEqual([0, 0, 100])
  // black [0, 0, 0] = [0, 0, 0]
  const convertBlackToHsl = rgbToHsl([0, 0, 0])
  expect(convertBlackToHsl).toEqual([0, 0, 0])
  // orange [255, 165, 0] = [39, 100, 50]
  const convertOrangeToHsl = rgbToHsl([255, 165, 0])
  expect(convertOrangeToHsl).toEqual([39, 100, 50])
  // green [0, 128, 0] = [120, 100, 25.1]
  const convertGreenToHsl = rgbToHsl([0, 128, 0])
  expect(convertGreenToHsl).toEqual([120, 100, 25.1])
  // pink [255, 192, 203] = [350, 100, 87.6]
  const convertPinkToHsl = rgbToHsl([255, 192, 203])
  expect(convertPinkToHsl).toEqual([350, 100, 87.6])
})

test('hslToRgb should convert an hsl color into its known rgb value', () => {
  // white [0, 0, 100] = [255, 255, 255]
  const convertWhiteToRgb = hslToRgb([0, 0, 100])
  expect(convertWhiteToRgb).toEqual([255, 255, 255])
  // black [0, 0, 0] = [0, 0, 0]
  const convertBlackToRgb = hslToRgb([0, 0, 0])
  expect(convertBlackToRgb).toEqual([0, 0, 0])
  // turquoise [174, 72.1, 56.5] = [64, 224, 208]
  const convertTurquoiseToRgb = hslToRgb([174, 72.1, 56.5])
  expect(convertTurquoiseToRgb).toEqual([64, 224, 208])
  // red [0, 100, 50] = [255, 0, 0]
  const convertRedToRgb = hslToRgb([0, 100, 50])
  expect(convertRedToRgb).toEqual([255, 0, 0])
  // grey [0, 0, 50.2] = [128, 128, 128]
  const convertGreyToRgb = hslToRgb([0, 0, 50.2])
  expect(convertGreyToRgb).toEqual([128, 128, 128])
})
