import _ from 'lodash'
import { ApolloError } from '@apollo/client'
import { GraphQLError } from 'graphql/error'
import { CombinedError } from 'urql'
import { useDestinationType } from './RequireConfig'
import { ErrorCode } from '../../schema'

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

export interface SimpleError {
  message: string
}

export interface KnownError extends GraphQLError, SimpleError {
  readonly path: ReadonlyArray<string | number>
  extensions: {
    code: ErrorCode
  }
}

export interface InputFieldError extends KnownError {
  extensions: {
    code: 'INVALID_INPUT_VALUE'
  }
}

export interface InvalidDestFieldValueError extends KnownError {
  extensions: {
    code: 'INVALID_DEST_FIELD_VALUE'
    fieldID: string
  }
}

function assertNever(x: never): void {
  console.log('unhandled error code', x)
}

function isKnownErrorCode(code: ErrorCode): code is ErrorCode {
  switch (code) {
    case 'INVALID_INPUT_VALUE':
      return true
    case 'INVALID_DEST_FIELD_VALUE':
      return true
    default:
      assertNever(code) // ensure we handle all error codes
      return false
  }
}

function isGraphQLError(err: unknown): err is GraphQLError {
  if (!err) return false
  if (!Object.prototype.hasOwnProperty.call(err, 'path')) return false
  if (!Object.prototype.hasOwnProperty.call(err, 'extensions')) return false
  return true
}

export function isKnownError(err: unknown): err is KnownError {
  if (!isGraphQLError(err)) return false
  if (!Object.prototype.hasOwnProperty.call(err.extensions, 'code'))
    return false

  return isKnownErrorCode(err.extensions.code as ErrorCode)
}
export function isDestFieldError(
  err: unknown,
): err is InvalidDestFieldValueError {
  if (!isKnownError(err)) return false
  return err.extensions.code === 'INVALID_DEST_FIELD_VALUE'
}
export function isInputFieldError(err: unknown): err is InputFieldError {
  if (!isKnownError(err)) return false
  return err.extensions.code === 'INVALID_INPUT_VALUE'
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
  errs: SimpleError[] | undefined | null,
): [InputFieldError[], SimpleError[]] {
  if (!errs) return [[], []]
  const inputFieldErrors = [] as InputFieldError[]
  const otherErrors = [] as SimpleError[]
  errs.forEach((err) => {
    if (!isInputFieldError(err)) {
      otherErrors.push(err)
      return
    }

    const fullPath = err.path.join('.')

    if (!filterPaths.includes(fullPath)) {
      otherErrors.push(err)
      return
    }

    inputFieldErrors.push(err)
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
): [InputFieldError | undefined, InvalidDestFieldValueError[], SimpleError[]] {
  const cfg = useDestinationType(destType) // need to call hook before conditional return
  if (!err) return [undefined, [], []]

  const [destTypeErrs, nonDestTypeErrs] = getInputFieldErrors(
    [destFieldPath + '.type'],
    err.graphQLErrors,
  )
  const destFieldErrs: InvalidDestFieldValueError[] = []
  const otherErrs: SimpleError[] = []

  nonDestTypeErrs.forEach((err) => {
    if (!isDestFieldError(err)) {
      otherErrs.push(err)
      return
    }

    const fullPath = err.path.join('.')
    if (fullPath !== destFieldPath) {
      otherErrs.push(err)
      return
    }

    const isReqField = cfg.requiredFields.some(
      (f) => f.fieldID === err.extensions.fieldID,
    )
    if (!isReqField) {
      otherErrs.push(err)
      return
    }

    destFieldErrs.push(err)
  })

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
