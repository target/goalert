import React, { useEffect, useState, useRef, useContext } from 'react'
import p from 'prop-types'
import { FormContext, FormContainerContext } from './context'
import { get, set, cloneDeep } from 'lodash'

// FormContainer handles grouping multiple FormFields.
// It works with the Form component to handle validation.
export function FormContainer(props) {
  const {
    value = {},
    mapValue = (value) => value,
    mapOnChangeValue = (value) => value,
    onChange: handleOnChange = () => {},
    optionalLabels,
    disabled: containerDisabled,
    removeFalseyIdxs,
    errors = [],
    children,
  } = props
  const [validationErrors, setValidationErrors] = useState([])
  const _fields = useRef({}).current

  const { disabled: formDisabled, addSubmitCheck } = useContext(FormContext)

  const addField = (fieldName, validate) => {
    if (!_fields[fieldName]) {
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

  const onSubmit = () => {
    const validate = (field) => {
      let err
      // find first error
      _fields[field].find((validate) => {
        err = validate(get(value, field))
        return err
      })
      if (err) err.field = field
      return err
    }
    const errs = Object.keys(_fields)
      .map(validate)
      .filter((e) => e)
    setValidationErrors(errs)
    if (errs.length) return false

    return true
  }

  useEffect(() => addSubmitCheck(onSubmit), [value])

  const onChange = (fieldName, e) => {
    // copy into a mutable object
    const oldValue = cloneDeep(value)

    let newValue = e
    if (e && e.target) newValue = e.target.value

    // remove idx from array if new value is null when fieldName includes index
    // e.g. don't set array to something like [3, null, 6, 2, 9]
    // if "array[1]" is null, but rather set to [3, 6, 2, 9]
    if (
      !newValue &&
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

      return handleOnChange(
        mapOnChangeValue(set(mapValue(oldValue), arrayPath, newArr)),
      )
    }

    return handleOnChange(
      mapOnChangeValue(set(mapValue(oldValue), fieldName, newValue)),
    )
  }

  const renderComponent = () => {
    return (
      <FormContainerContext.Provider
        value={{
          value: mapValue(value),
          disabled: formDisabled || containerDisabled,
          errors: validationErrors.concat(errors),
          onChange,
          addField,
          optionalLabels,
        }}
      >
        {children}
      </FormContainerContext.Provider>
    )
  }

  return <FormContext.Consumer>{renderComponent}</FormContext.Consumer>
}

FormContainer.propTypes = {
  value: p.object,
  children: p.node,
  errors: p.arrayOf(
    p.shape({
      message: p.string,
      field: p.string,
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
