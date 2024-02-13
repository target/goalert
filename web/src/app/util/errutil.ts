import _ from 'lodash'
import { ApolloError } from '@apollo/client'
import { GraphQLError } from 'graphql/error'
import { CombinedError } from 'urql'
import {
  INVALID_DESTINATION_FIELD_VALUE,
  INVALID_DESTINATION_TYPE,
} from '../../errors.d'
import { useDestinationType } from './RequireConfig'

const mapName = (name: string): string => _.camelCase(name).replace(/Id$/, 'ID')

// stripMessage will filter out any details attributes from an error message
const stripMessage = (msg: string): string => msg.split(';')[0]

// parseDetails will parse out any details information from an error message
const parseDetails = (msg: string): { [x: string]: string } => {
  const parts = msg.split(';').slice(1)
  if (!parts.length) return {}
  const details: { [x: string]: string } = {}
  parts.forEach((p) => {
    const keyVal = p.split('=')
    details[keyVal[0].trim()] = keyVal.slice(1).join('=').trim()
  })
  return details
}

// nonFieldErrors will return a flat list of non-field errors (if any) from a graphQL error.
//
// All returned errors should have a `message` property.
export function nonFieldErrors(err?: ApolloError | CombinedError): Error[] {
  if (!err) return []
  if (!err.graphQLErrors || !err.graphQLErrors.length) return [err]

  return (err.graphQLErrors as GraphQLError[]).filter(
    (err) =>
      !err.extensions ||
      !(err.extensions.isFieldError || err.extensions.isMultiFieldError),
  )
}

export type SimpleError = {
  message: string
}

export type InputFieldError = {
  message: string
  path: string
}

/**
 * getInputFieldErrors returns a list of input field errors and other errors from a CombinedError.
 * Any errors that are not input field errors (or are not in the filterPaths list) will be returned as other errors.
 *
 * @param filterPaths - a list of paths to filter errors by
 * @param err - the CombinedError to filter
 */
export function getInputFieldErrors(
  filterPaths: string[],
  err: CombinedError,
): [InputFieldError[], SimpleError[]] {
  const inputFieldErrors = [] as InputFieldError[]
  const otherErrors = [] as SimpleError[]
  err.graphQLErrors.forEach((err) => {
    if (!err.path) {
      otherErrors.push({ message: err.message })
      return
    }
    const code = err.extensions?.code
    if (
      // only support known error codes
      code !== INVALID_DESTINATION_TYPE &&
      code !== INVALID_DESTINATION_FIELD_VALUE
    ) {
      otherErrors.push({ message: err.message })
      return
    }

    const fullPath = err.path.join('.')

    if (!filterPaths.includes(fullPath)) {
      otherErrors.push({ message: err.message })
      return
    }

    inputFieldErrors.push({ message: err.message, path: fullPath })
  })

  return [inputFieldErrors, otherErrors]
}

/**
 * useErrorsForDest returns the errors for a destination type and field path from a CombinedError.
 * The first return value is the error for the destination type, if any.
 * The second return value is a list of errors for the destination fields, if any.
 * The third return value is a list of other errors, if any.
 */
export function useErrorsForDest(
  err: CombinedError | undefined | null,
  destType: string,
  destFieldPath: string, // the path of the DestinationInput field
): [SimpleError | undefined, InputFieldError[], SimpleError[]] {
  const cfg = useDestinationType(destType) // need to call hook before conditional return
  if (!err) return [undefined, [], []]

  const [destTypeErrs] = getInputFieldErrors([destFieldPath + '.type'], err)

  const paths = cfg.requiredFields.map(
    (f) => `${destFieldPath}.values.${f.fieldID}`,
  )

  const [destFieldErrs, otherErrs] = getInputFieldErrors(paths, err)

  return [destTypeErrs[0] || undefined, destFieldErrs, otherErrs]
}

export interface FieldError extends Error {
  field: string
  details: { [x: string]: string }
  path: GraphQLError['path']
}

function isFieldError(e: Error | FieldError): e is FieldError {
  return !!(e as FieldError).field
}

interface RawFieldError extends Error {
  fieldName: string
}
// fieldErrors will return a flat list of field errors (if any) from a graphQL error.
//
// All returned errors will be of the format {field, message}
export function fieldErrors(err?: ApolloError | CombinedError): FieldError[] {
  if (!err) return []
  if (!err.graphQLErrors) return []

  const errs = (err.graphQLErrors as GraphQLError[])
    .filter(
      (err) =>
        err.extensions &&
        (err.extensions.isFieldError || err.extensions.isMultiFieldError),
    )
    .map((err) => {
      if (err.extensions?.isMultiFieldError) {
        return (err.extensions.fieldErrors as RawFieldError[]).map((e) => ({
          field: e.fieldName.split('.').map(mapName).join('.'),
          message: stripMessage(e.message),
          details: parseDetails(e.message),
          path: err.path,
          name: 'FieldError',
        }))
      }

      return {
        field: (err.extensions?.fieldName as string)
          .split('.')
          .map(mapName)
          .join('.'),
        message: stripMessage(err.message),
        details: parseDetails(err.message),
        path: err.path,
        name: 'FieldError',
      }
    })

  return errs.flat()
}

// allErrors will return a flat list of all errors in the graphQL error.
export function allErrors(
  err?: ApolloError | CombinedError,
): (FieldError | Error)[] {
  if (!err) return []
  return nonFieldErrors(err).concat(fieldErrors(err))
}

// byPath will group errors by their path name.
export function errorsByPath(err: ApolloError): { [x: string]: Error[] } {
  return _.groupBy(allErrors(err), (e: Error | FieldError) =>
    (isFieldError(e) && e.path ? e.path : []).join('.'),
  )
}
