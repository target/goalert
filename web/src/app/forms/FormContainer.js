import React, { useState, useRef } from 'react'
import p from 'prop-types'
import MountWatcher from '../util/MountWatcher'

import { FormContext, FormContainerContext } from './context'
import { get, set, cloneDeep } from 'lodash'

// FormContainer handles grouping multiple FormFields.
// It works with the Form component to handle validation.
export function FormContainer(props) {
  const [validationErrors, setValidationErrors] = useState([])
  const _unregister = useRef()
  const _fields = useRef({})

  const addField = (fieldName, validate) => {
    if (!_fields.current[fieldName]) {
      _fields.current[fieldName] = []
    }
    _fields.current[fieldName].push(validate)

    return () => {
      _fields.current[fieldName] = _fields.current[fieldName].filter(
        (v) => v !== validate,
      )
      if (_fields.current[fieldName].length === 0) {
        delete _fields.current[fieldName]
      }
    }
  }

  const onSubmit = () => {
    const validate = (field) => {
      let err
      // find first error
      _fields.current[field].find((validate) => {
        err = validate(get(props.value, field))

        return err
      })
      if (err) err.field = field
      return err
    }
    const validationErrors = Object.keys(_fields.current)
      .map(validate)
      .filter((e) => e)
    setValidationErrors(validationErrors)
    if (validationErrors.length) return false

    return true
  }

  const contextOnChange = (fieldName, e) => {
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

  const renderComponent = ({ disabled: formDisabled, addSubmitCheck }) => {
    const {
      value,
      mapValue,
      optionalLabels,
      disabled: containerDisabled,
    } = props

    return (
      <MountWatcher
        onMount={() => {
          _unregister.current = () => addSubmitCheck(onSubmit)
        }}
        onUnmount={() => {
          _unregister.current()
        }}
      >
        <FormContainerContext.Provider
          value={{
            value: mapValue(value),
            disabled: formDisabled || containerDisabled,
            errors: validationErrors.concat(props.errors),
            onChange: contextOnChange,
            addField: addField,
            optionalLabels: optionalLabels,
          }}
        >
          {props.children}
        </FormContainerContext.Provider>
      </MountWatcher>
    )
  }

  return <FormContext.Consumer>{renderComponent}</FormContext.Consumer>
}

FormContainer.propTypes = {
  children: p.node,
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

FormContainer.defaultProps = {
  errors: [],
  value: {},
  onChange: () => {},

  mapValue: (value) => value,
  mapOnChangeValue: (value) => value,
}
