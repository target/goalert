/**
 * INVALID_DESTINATION_TYPE is returned when the selected destination type is not valid, or is not allowed.
 */
export const INVALID_DESTINATION_TYPE = 'INVALID_DESTINATION_TYPE'

/**
 * INVALID_DESTINATION_FIELD_VALUE is returned when the value of a field on a destination is invalid.
 */
export const INVALID_DESTINATION_FIELD_VALUE = 'INVALID_DESTINATION_FIELD_VALUE'

type KnownErrorCode = INVALID_DESTINATION_TYPE | INVALID_DESTINATION_FIELD_VALUE

export type InvalidDestTypeError = {
  message: string
  path: readonly (string | number)[]
  extensions: {
    code: INVALID_DESTINATION_TYPE
  }
}

export type InvalidFieldValueError = {
  message: string
  path: readonly (string | number)[]
  extensions: {
    code: INVALID_DESTINATION_FIELD_VALUE
  }
}
