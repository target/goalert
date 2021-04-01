import { intersection } from 'lodash-es'

// oneOfShape requires that one of the props defined
// is provided.
export function oneOfShape(config) {
  return (props, propName, ...args) => {
    const keys = Object.keys(config)
    const overlap = intersection(Object.keys(props), keys)
    if (overlap.length === 0) {
      return new Error(`One of [${keys.join(', ')}] is required.`)
    }
    if (overlap.length > 1) {
      return new Error(
        `Only one of [${keys.join(
          ', ',
        )}] should be provided, but found [${overlap.join(', ')}].`,
      )
    }

    const key = overlap[0]
    return config[key](props, key, ...args)
  }
}
