import { GraphQLError } from 'graphql'
import { ErrorCode } from '../../schema'

export interface BaseError {
  message: string
}

export interface KnownError extends GraphQLError, BaseError {
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
