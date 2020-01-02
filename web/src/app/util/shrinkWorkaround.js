import _ from 'lodash-es'
// fixes bug in material where non-empty values
// fail to trigger the label text to shrink.
//
// Usage:
// const InputProps = {
//  ...otherInputProps,
//  ...shrinkWorkaround(value)
// }
export default function shrinkWorkaround(value) {
  if (_.isEmpty(value) && !_.isNumber(value)) return {}
  return { shrink: true }
}
