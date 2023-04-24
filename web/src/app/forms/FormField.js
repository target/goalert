import React, { useContext, useEffect } from 'react'
import p from 'prop-types'
import FormControl from '@mui/material/FormControl'
import FormHelperText from '@mui/material/FormHelperText'
import FormLabel from '@mui/material/FormLabel'
import { get, isEmpty, startCase } from 'lodash'
import shrinkWorkaround from '../util/shrinkWorkaround'
import AppLink from '../util/AppLink'
import { FormContainerContext } from './context'
import { Grid } from '@mui/material'

export function FormField(props) {
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
    component: Component,
    render,
    fieldName: _fieldName,
    formLabel,
    required,
    validate = () => {},
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
    charCount,
    ...otherFieldProps
  } = props

  const fieldName = _fieldName || name

  const validateField = (value) => {
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

  const fieldProps = {
    ...otherFieldProps,
    name,
    required,
    disabled: containerDisabled || fieldDisabled,
    error: errors.find((err) => err.field === (errorName || fieldName)),
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

  let getValueOf = (e) => (e && e.target ? e.target.value : e)
  if (checkbox) {
    fieldProps.checked = fieldProps.value
    fieldProps.value = fieldProps.value ? 'true' : 'false'
    getValueOf = (e) => e.target.checked
  } else if (otherFieldProps.type === 'number') {
    fieldProps.label = label
    fieldProps.value = fieldProps.value.toString()
    fieldProps.InputLabelProps = InputLabelProps
    getValueOf = (e) =>
      float ? parseFloat(e.target.value) : parseInt(e.target.value, 10)
  } else {
    fieldProps.label = label
    fieldProps.InputLabelProps = InputLabelProps
  }

  fieldProps.onChange = (_value) => {
    let newValue = getValueOf(_value)
    if (fieldProps.type === 'number' && typeof fieldProps.min === 'number')
      newValue = Math.max(fieldProps.min, newValue)
    if (fieldProps.type === 'number' && typeof fieldProps.max === 'number')
      newValue = Math.min(fieldProps.max, newValue)
    onChange(fieldName, mapOnChangeValue(newValue, value))
  }

  // wraps hints/errors within a grid containing character counter to align horizontally
  function charCountWrapper(component, count) {
    return (
      <Grid container spacing={2}>
        <Grid item xs={10}>
          {component}
        </Grid>
        <Grid item xs={2}>
          <FormHelperText style={{ textAlign: 'right' }}>
            {value.description.length}/{count}
          </FormHelperText>
        </Grid>
      </Grid>
    )
  }

  function renderFormHelperText(error, hint, count) {
    // handle optional count parameter
    if (count === undefined) {
      count = 0
    }
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
        const errorText = (
          <FormHelperText>
            {error.message.replace(/^./, (str) => str.toUpperCase())}
          </FormHelperText>
        )
        if (count) {
          return charCountWrapper(errorText, count)
        }
        return errorText
      }
    }

    if (hint) {
      if (count) {
        return charCountWrapper(<FormHelperText>{hint}</FormHelperText>, count)
      }
      return <FormHelperText>{hint}</FormHelperText>
    }

    return null
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
      {renderFormHelperText(fieldProps.error, fieldProps.hint, charCount)}
    </FormControl>
  )
}
FormField.propTypes = {
  // pass select dropdown items as children
  children: p.node,

  // one of component or render must be provided
  component: p.any,
  render: p.func,

  // mapValue can be used to map a value before it's passed to the form component
  mapValue: p.func,

  // mapOnChangeValue can be used to map a changed value from the component, before it's
  // passed to the parent form's state.
  mapOnChangeValue: p.func,

  // Adjusts props for usage with a Checkbox component.
  checkbox: p.bool,

  // Allows entering decimal number into a numeric field.
  float: p.bool,

  // fieldName specifies the field used for
  // checking errors, change handlers, and value.
  //
  // If unset, it defaults to `name`.
  name: p.string.isRequired,
  fieldName: p.string,

  // min and max values specify the range to clamp a int value
  // expects an ISO timestamp, if string
  min: p.oneOfType([p.number, p.string]),
  max: p.oneOfType([p.number, p.string]),

  // used if name is set,
  // but the error name is different from graphql responses
  errorName: p.string,

  // label above form component
  label: p.node,
  formLabel: p.bool, // use formLabel instead of label if true

  // required indicates the field may not be left blank.
  required: p.bool,

  // validate can be used to provide client-side validation of a
  // field.
  validate: p.func,

  // a hint for the user on a form field. errors take priority
  hint: p.node,

  // disable the form helper text for errors.
  noError: p.bool,

  step: p.oneOfType([p.number, p.string]),

  InputProps: p.object,

  disabled: p.bool,

  multiline: p.bool,
  rows: p.number,
  autoComplete: p.string,

  charCount: p.number,

  fullWidth: p.bool,

  placeholder: p.string,

  type: p.string,
  select: p.bool,
  timeZone: p.string,

  userID: p.string,

  value: p.oneOfType([p.string, p.arrayOf(p.string)]),

  multiple: p.bool,

  options: p.shape({
    label: p.string,
    value: p.string,
  }),
}
