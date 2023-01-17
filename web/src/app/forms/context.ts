import React from 'react'
import { FieldError } from '../util/errutil'
import { MapFuncType, Validate, Value } from './FormField'

type AddField = (field?: string, validate?: Validate) => void

export const FormContainerContext = React.createContext({
  onChange: (() => {}) as (
    fieldName: string,
    mapOnChangeValue?: MapFuncType,
  ) => void,
  disabled: false as boolean,
  errors: [] as FieldError[],
  value: {} as Value,
  addField: (() => () => {}) as AddField,
  optionalLabels: false as boolean,
})
FormContainerContext.displayName = 'FormContainerContext'

export const FormContext = React.createContext({
  disabled: false,
  addSubmitCheck: () => () => {},
})

FormContext.displayName = 'FormContext'
