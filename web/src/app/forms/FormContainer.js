import React, { useState } from 'react'
import PropTypes from 'prop-types'
import MountWatcher from '../util/MountWatcher'

import { FormContext, FormContainerContext } from './context'
import { get, set, cloneDeep } from 'lodash'

// FormContainer handles grouping multiple FormFields.
// It works with the Form component to handle validation.

export function FormContainer(props) {
  const [state, setState] = useState({})
  _fields = {}
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
    setState({ validationErrors })
    if (validationErrors.length) return false

    return true
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

  function renderComponent({ disabled: formDisabled, addSubmitCheck }) {
    const {
      value,
      mapValue,
      optionalLabels,
      disabled: containerDisabled,
    } = props

    return (
      <MountWatcher
        onMount={() => {
          _unregister = addSubmitCheck(onSubmit)
        }}
        onUnmount={() => {
          _unregister()
        }}
      >
        <FormContainerContext.Provider
          value={{
            value: mapValue(value),
            disabled: formDisabled || containerDisabled,
            errors: validationErrors.concat(props.errors),
            onChange: onChange,
            addField: addField,
            optionalLabels: optionalLabels,
          }}
        >
          {props.children}
          <h1>GHJGHGHFGHFJFFGHFGHFJGHFGHFJ</h1>
        </FormContainerContext.Provider>
      </MountWatcher>
    )
  }
}

FormContainer.propTypes = {
  value: PropTypes.object,
  errors: PropTypes.arrayOf(
    PropTypes.shape({
      message: PropTypes.string.isRequired,
      field: PropTypes.string.isRequired,
      helpLink: PropTypes.string,
    }),
  ),

  onChange: PropTypes.func,
  disabled: PropTypes.bool,

  mapValue: PropTypes.func,
  mapOnChangeValue: PropTypes.func,

  // If true, will render optional fields with `(optional)` appended to the label.
  // In addition, required fields will not be appended with `*`.
  optionalLabels: PropTypes.bool,

  // Enables functionality to remove an incoming value at it's index from
  // an array field if the new value is falsey.
  removeFalseyIdxs: PropTypes.bool,
}
