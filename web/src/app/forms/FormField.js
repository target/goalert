import React, { useContext, useEffect } from 'react'
import p from 'prop-types'
import FormControl from '@material-ui/core/FormControl'
import FormHelperText from '@material-ui/core/FormHelperText'
import FormLabel from '@material-ui/core/FormLabel'
import { get, isEmpty, startCase } from 'lodash'
import shrinkWorkaround from '../util/shrinkWorkaround'
import AppLink from '../util/AppLink'
import { FormContainerContext } from './context'

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
    validate,
    disabled: fieldDisabled,
    hint,
    label: _label,
    InputLabelProps: _inputProps,
    mapValue,
    mapOnChangeValue,
    min,
    max,
    checkbox,
    float,
    ...otherFieldProps
  } = props

  const baseLabel = typeof _label === 'string' ? _label : startCase(name)
  const label =
    !required && optionalLabels ? baseLabel + ' (optional)' : baseLabel

  const fieldName = _fieldName || name

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
  }, [])

  function renderFormHelperText(error, hint) {
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

    if (hint) {
      return <FormHelperText>{hint}</FormHelperText>
    }

    return null
  }

  function renderContent() {
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
          label={formLabel ? null : fieldProps.label}
        >
          {fieldProps.children}
        </Component>
        {!noError && renderFormHelperText(fieldProps.error, fieldProps.hint)}
      </FormControl>
    )
  }

  return (
    <FormContainerContext.Consumer>
      {renderContent}
    </FormContainerContext.Consumer>
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
  autoComplete: p.string,

  fullWidth: p.bool,

  placeholder: p.string,

  type: p.string,
  select: p.bool,
  timeZone: p.string,
}

FormField.defaultProps = {
  validate: () => {},
  mapValue: (value) => value,
  mapOnChangeValue: (value) => value,
}
