import React, { useEffect, useState, useRef, useContext } from 'react'
import p from 'prop-types'
import { FormContext, FormContainerContext } from './context'
import { get, set, cloneDeep } from 'lodash'

// FormContainer handles grouping multiple FormFields.
// It works with the Form component to handle validation.

export function FormContainer(props) {
  const [persistentValidationErrors, setPersistentValidationErrors] = useState(
    [],
  )
  const _fields = useRef().current
  const { value, mapValue, optionalLabels, disabled: containerDisabled } = props
  const { disabled: formDisabled, addSubmitCheck } = useContext(FormContext)

  function onSubmit() {
    function validate(field) {
      let err
      // find first error
      _fields[field].find((validate) => {
        err = validate(get(props.value, field))
        return err
      })
      if (err) err.field = field
      return err
    }
    const validationErrors = Object.keys(_fields)
      .map(validate)
      .filter((e) => e)
    setPersistentValidationErrors({ validationErrors })
    if (validationErrors.length) return false

    return true
  }

  useEffect(() => addSubmitCheck(onSubmit), [])

  function addField(fieldName, validate) {
    if (_fields[fieldName]) {
      _fields[fieldName] = []
    }
    _fields[fieldName].push(validate)

    return () => {
      _fields[fieldName] = _fields[fieldName].filter((v) => v !== validate)
      if (_fields[fieldName].length === 0) {
        delete _fields[fieldName]
      }
    }
  }

  function onChange(fieldName, e) {
    const {
      mapValue,
      mapOnChangeValue,
      value: _value,
      removeFalseyIdxs,
    } = props

    // copy into a mutable object
    const oldValue = cloneDeep(_value)

    let value = e
    if (e && e.target) value = e.target.value

    // remove idx from array if new value is null when fieldName includes index
    // e.g. don't set array to something like [3, null, 6, 2, 9]
    // if "array[1]" is null, but rather set to [3, 6, 2, 9]
    if (
      !value &&
      fieldName.charAt(fieldName.length - 1) === ']' &&
      removeFalseyIdxs
    ) {
      const arrayPath = fieldName.substring(0, fieldName.lastIndexOf('['))
      const idx = fieldName.substring(
        fieldName.lastIndexOf('[') + 1,
        fieldName.lastIndexOf(']'),
      )

      const newArr = get(oldValue, arrayPath, []).filter((_, i) => {
        return i !== parseInt(idx, 10)
      })

      return props.onChange(
        mapOnChangeValue(set(mapValue(oldValue), arrayPath, newArr)),
      )
    }

    return props.onChange(
      mapOnChangeValue(set(mapValue(oldValue), fieldName, value)),
    )
  }

  return (
    <FormContainerContext.Provider
      value={{
        value: mapValue(value),
        disabled: formDisabled || containerDisabled,
        errors: persistentValidationErrors.concat(props.errors),
        onChange: onChange,
        addField: addField,
        optionalLabels: optionalLabels,
      }}
    >
      {props.children}
    </FormContainerContext.Provider>
  )
}

FormContainer.propTypes = {
  value: p.object,
  errors: p.arrayOf(
    p.shape({
      message: p.string.isRequired,
      field: p.string.isRequired,
      helpLink: p.string,
    }),
  ),

  onChange: p.func,
  disabled: p.bool,

  mapValue: p.func,
  mapOnChangeValue: p.func,

  // If true, will render optional fields with `(optional)` appended to the label.
  // In addition, required fields will not be appended with `*`.
  optionalLabels: p.bool,

  // Enables functionality to remove an incoming value at it's index from
  // an array field if the new value is falsey.
  removeFalseyIdxs: p.bool,
}
