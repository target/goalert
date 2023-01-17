import React, { ElementType, useContext, useEffect, ReactNode } from 'react'
import FormControl from '@mui/material/FormControl'
import FormHelperText from '@mui/material/FormHelperText'
import FormLabel from '@mui/material/FormLabel'
import { get, isEmpty, startCase } from 'lodash'
import shrinkWorkaround from '../util/shrinkWorkaround'
import AppLink from '../util/AppLink'
import { FormContainerContext } from './context'
import { InputLabelProps, InputProps } from '@mui/material'
import { FieldError } from '../util/errutil'

// children components will define the value types
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type MapFuncType = (value: any, value2?: any) => any
export type Value = unknown
export type OnChange = (
  fieldName: string | string[],
  mapOnChangeValue?: MapFuncType,
) => void

type ValidateResult = boolean | Error | void | null
export type Validate = (value: Value) => ValidateResult

type Option = {
  label: string
  value: string
}

interface CustomError {
  message: string
  field: string
  helpLink?: string
}

interface FormFieldProps {
  // pass select dropdown items as children
  children?: ReactNode | ReactNode[]

  // one of component or render must be provided
  component?: ElementType
  render?: (props: Partial<FormFieldProps>) => JSX.Element

  // mapValue can be used to map a value before it's passed to the form component
  mapValue?: MapFuncType

  // mapOnChangeValue can be used to map a changed value from the component, before it's
  // passed to the parent form's state.
  mapOnChangeValue?: MapFuncType

  // Adjusts props for usage with a Checkbox component.
  checkbox?: boolean

  // Allows entering decimal number into a numeric field.
  float?: boolean

  // fieldName specifies the field used for
  // checking errors, change handlers, and value.
  //
  // If unset, it defaults to `name`.
  name: string
  fieldName?: string

  // min and max values specify the range to clamp a int value
  // expects an ISO timestamp, if string
  min?: number | string
  max?: number | string

  // used if name is set,
  // but the error name is different from graphql responses
  errorName?: string

  // label above form component
  label?: ReactNode
  formLabel?: boolean // use formLabel instead of label if true

  // required indicates the field may not be left blank.
  required?: boolean

  // validate can be used to provide client-side validation of a
  // field.
  validate?: Validate

  // a hint for the user on a form field. errors take priority
  hint?: ReactNode

  // disable the form helper text for errors.
  noError?: boolean

  step?: number | string

  InputProps?: InputProps
  InputLabelProps?: InputLabelProps

  disabled?: boolean

  multiline?: boolean
  autoComplete?: string

  fullWidth?: boolean

  placeholder?: string

  type?: string
  select?: boolean
  timeZone?: string

  userID?: string

  value?: Value

  multiple?: boolean

  options?: Option | Option[]

  // FieldProps - todo: use seperate type that inherits needed props from above?
  error?: FieldError | Error | CustomError
  checked?: boolean
  onChange?: OnChange
}

export function FormField(props: FormFieldProps): JSX.Element {
  const {
    errors,
    value,
    onChange,
    addField,
    disabled: containerDisabled,
    optionalLabels,
  } = useContext(FormContainerContext)

  const {
    errorName,
    name,
    noError,
    component: Component = React.Fragment,
    render,
    fieldName: _fieldName,
    formLabel,
    required,
    validate = (value: unknown): boolean => Boolean(value),
    disabled: fieldDisabled,
    hint,
    label: _label,
    InputLabelProps: _inputProps,
    mapValue = (value) => value,
    mapOnChangeValue = (value) => value,
    min,
    max,
    checkbox,
    float,
    ...otherFieldProps
  } = props

  const fieldName = _fieldName || name

  const validateField = (value: Value): ValidateResult => {
    if (
      required &&
      !['boolean', 'number'].includes(typeof value) &&
      isEmpty(value)
    ) {
      return new Error('Required field.')
    }

    return validate(value)
  }

  useEffect(() => {
    return addField(fieldName, validateField)
  }, [required])

  const baseLabel = typeof _label === 'string' ? _label : startCase(name)
  const label =
    !required && optionalLabels ? baseLabel + ' (optional)' : baseLabel

  const fieldProps: Partial<FormFieldProps> = {
    ...otherFieldProps,
    name,
    required,
    disabled: containerDisabled || fieldDisabled,
    error: errors.find(
      (err) => err.field === (errorName || fieldName),
    ) as CustomError,
    hint,
    value: mapValue(get(value, fieldName), value),
    min,
    max,
    float,
  }

  const InputLabelProps = {
    required: required && !optionalLabels,
    ...shrinkWorkaround(props.value),
    ..._inputProps,
  }

  type VE = React.ChangeEvent<HTMLInputElement> | string | string[]
  // mutable function
  let getValueOf = (e: VE): unknown => {
    if (typeof e === 'string' || e instanceof Array) {
      return e
    }
    return e.target.value
  }

  if (checkbox) {
    fieldProps.checked = fieldProps.value as boolean
    fieldProps.value = fieldProps.value ? 'true' : 'false'
    getValueOf = (e: VE) => {
      if (typeof e === 'string' || e instanceof Array) {
        return e
      }
      return e.target.checked
    }
  } else if (otherFieldProps.type === 'number') {
    fieldProps.label = label
    fieldProps.value = (fieldProps.value as number).toString()
    fieldProps.InputLabelProps = InputLabelProps
    getValueOf = (e: VE) => {
      if (typeof e === 'string' || e instanceof Array) {
        return e
      }
      const v = e.target.value
      return float ? parseFloat(v) : parseInt(v, 10)
    }
  } else {
    fieldProps.label = label
    fieldProps.InputLabelProps = InputLabelProps
  }

  fieldProps.onChange = (_value) => {
    let newValue = getValueOf(_value)
    if (fieldProps.type === 'number' && typeof fieldProps.min === 'number')
      newValue = Math.max(fieldProps.min, newValue as number)
    if (fieldProps.type === 'number' && typeof fieldProps.max === 'number')
      newValue = Math.min(fieldProps.max, newValue as number)

    onChange(fieldName, mapOnChangeValue(newValue, value))
  }

  function renderFormHelperText(
    error: CustomError,
    hint: ReactNode,
  ): ReactNode {
    if (!error && !hint) return

    if (!noError) {
      if (error?.helpLink) {
        return (
          <FormHelperText>
            <AppLink to={error.helpLink} newTab data-cy='error-help-link'>
              {error.message.replace(/^./, (str) => str.toUpperCase())}
            </AppLink>
          </FormHelperText>
        )
      }

      if (error?.message) {
        return (
          <FormHelperText>
            {error.message.replace(/^./, (str) => str.toUpperCase())}
          </FormHelperText>
        )
      }
    }

    if (hint) {
      return <FormHelperText>{hint}</FormHelperText>
    }
  }

  if (render) return render(fieldProps)

  return (
    <FormControl
      fullWidth={fieldProps.fullWidth}
      error={Boolean(fieldProps.error)}
    >
      {formLabel && (
        <FormLabel style={{ paddingBottom: '0.5em' }}>{_label}</FormLabel>
      )}
      <Component
        {...fieldProps}
        error={checkbox ? undefined : Boolean(fieldProps.error)}
        // NOTE: empty string label leaves gap in outlined field; fallback to undefined instead
        label={(!formLabel && fieldProps.label) || undefined}
      >
        {fieldProps.children}
      </Component>
      {renderFormHelperText(fieldProps.error as CustomError, fieldProps.hint)}
    </FormControl>
  )
}
