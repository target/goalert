import _ from 'lodash'

// shrinkWorkaround fixes bug in material where non-empty values
// fail to trigger the label text to shrink.
//
// Usage:
// const InputProps = {
//  ...otherInputProps,
//  ...shrinkWorkaround(value)
// }
export default function shrinkWorkaround(value: string | number): {
  shrink?: boolean
} {
  if (_.isEmpty(value) && !_.isNumber(value)) return {}
  return { shrink: true }
}
