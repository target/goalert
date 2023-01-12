import React from 'react'
import { FieldError } from '../util/errutil'
import { MapOnChangeValue } from './FormField'

type AddField = (
  field?: string,
  validate?: (value: unknown) => boolean | Error | void,
) => void

export const FormContainerContext = React.createContext({
  onChange: (() => {}) as (
    fieldName: string,
    mapOnChangeValue?: MapOnChangeValue,
  ) => void,
  disabled: false as boolean,
  errors: [] as FieldError[],
  value: {} as unknown,
  addField: (() => () => {}) as AddField,
  optionalLabels: false as boolean,
})
FormContainerContext.displayName = 'FormContainerContext'

export const FormContext = React.createContext({
  disabled: false,
  addSubmitCheck: () => () => {},
})

FormContext.displayName = 'FormContext'
