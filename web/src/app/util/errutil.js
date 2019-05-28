import { camelCase } from 'lodash-es'

// allErrors will return a flat list of all errors in the graphQL error.
export function allErrors(err) {
  const errs = fieldErrors(err).concat(nonFieldErrors(err))
  return [].concat(...errs)
}

const mapName = name => camelCase(name).replace(/Id$/, 'ID')
const stripMessage = msg => msg.split(';')[0]
const stripDetails = msg => {
  const parts = msg.split(';').slice(1)
  if (!parts.length) return null
  const details = {}
  parts.forEach(p => {
    const keyVal = p.split('=')
    details[keyVal[0].trim()] = keyVal
      .slice(1)
      .join('=')
      .trim()
  })
  return details
}

// fieldErrors will return a flat list of field errors (if any) from a graphQL error.
//
// All returned errors will be of the format {field, message}
export function fieldErrors(err) {
  if (!err) return []
  if (!err.graphQLErrors) return []

  const errs = err.graphQLErrors
    .filter(
      err =>
        err.extensions &&
        (err.extensions.isFieldError || err.extensions.isMultiFieldError),
    )
    .map(err => {
      if (err.extensions.isMultiFieldError) {
        return err.extensions.fieldErrors.map(e => ({
          field: e.fieldName
            .split('.')
            .map(mapName)
            .join('.'),
          message: stripMessage(e.message),
          details: stripDetails(e.message),
        }))
      }

      return {
        field: err.extensions.fieldName
          .split('.')
          .map(mapName)
          .join('.'),
        message: stripMessage(err.message),
        details: stripDetails(err.message),
      }
    })

  return [].concat(...errs)
}

// nonFieldErrors will return a flat list of non-field errors (if any) from a graphQL error.
//
// All returned errors should have a `message` property.
export function nonFieldErrors(err) {
  if (!err) return []
  if (!err.graphQLErrors || !err.graphQLErrors.length) return [err]

  return err.graphQLErrors.filter(
    err =>
      !err.extensions ||
      !(err.extensions.isFieldError || err.extensions.isMultiFieldError),
  )
}
