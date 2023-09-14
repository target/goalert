import React from 'react'
import MountWatcher from '../util/MountWatcher'

import { FormContext, FormContainerContext } from './context'
import { get, set, cloneDeep } from 'lodash'

interface Error {
  message: string
  field: string
  helpLink?: string
}

interface FormContainerProps {
  value?: Object | undefined;
  errors?: Error[] | any;
  onChange?: any;
  disabled?: boolean | any;
  mapValue?: any;
  mapOnChangeValue?: any;
  // If true, will render optional fields with `(optional)` appended to the label.
  // In addition, required fields will not be appended with `*`.
  optionalLabels?: boolean;
  // Enables functionality to remove an incoming value at it's index from
  // an array field if the new value is falsey.
  removeFalseyIdxs?: boolean;
}

interface Formcheck {
  disabled: boolean, 
  addSubmitCheck: any 
}

// FormContainer handles grouping multiple FormFields.
// It works with the Form component to handle validation.
export class FormContainer extends React.PureComponent <FormContainerProps> {
  static defaultProps: Partial<FormContainerProps> = {
    errors: [],
    value: {},
    onChange: () => {},
    mapValue: (value: any) => value,
    mapOnChangeValue: (value: any) => value,
  };

  state = {
    validationErrors: [],
  }

  _fields: any = {}

  addField = (fieldName: string, validate: any) => {
    if (!this._fields[fieldName]) {
      this._fields[fieldName] = []
    }
    this._fields[fieldName].push(validate)

    return () => {
      this._fields[fieldName] = this._fields[fieldName].filter(
        (v: any) => v !== validate,
      )
      if (this._fields[fieldName].length === 0) {
        delete this._fields[fieldName]
      }
    }
  }

  onSubmit = () => {
    const validate = (field: any) => {
      let err: Error
      // find first error
      err = this._fields[field].find((validate: any) => {
        // console.log to check value
        console.log("ERRORForm",err)
        err = validate(get(this.props.value, field))
        console.log("ERROR", err)
        return err
      })
      
      if (err) { 
        err.field = field 
      }
      return err
    }
    const validationErrors = Object.keys(this._fields)
      .map(validate)
      .filter((e) => e)
    this.setState({ validationErrors })
    if (validationErrors.length) return false

    return true
  }

  onChange = (fieldName: string, e: any) => {
    const {
      mapValue,
      mapOnChangeValue,
      value: _value,
      removeFalseyIdxs,
    } = this.props

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

      const newArr = get(oldValue, arrayPath, []).filter((_: undefined, i: Number) => {
        return i !== parseInt(idx, 10)
      })

      return this.props.onChange(
        mapOnChangeValue(set(mapValue(oldValue), arrayPath, newArr)),
      )
    }

    return this.props.onChange(
      mapOnChangeValue(set(mapValue(oldValue), fieldName, value)),
    )
  }
  _unregister: any

  render() {
    return <FormContext.Consumer>{this.renderComponent}</FormContext.Consumer>
  }

  renderComponent = (formcheck: Formcheck) => {
    const { disabled: formDisabled, addSubmitCheck }:Formcheck= formcheck

    const {
      value,
      mapValue,
      optionalLabels,
      disabled: containerDisabled,
    } = this.props

    return (
      <MountWatcher
        onMount={() => {
          this._unregister = addSubmitCheck(this.onSubmit)
        }}
        onUnmount={() => {
          this._unregister()
        }}
      >
        <FormContainerContext.Provider
          value={{
            value: mapValue(value),
            disabled: formDisabled || containerDisabled,
            errors: this.state.validationErrors.concat(this.props.errors),
            onChange: this.onChange,
            addField: this.addField,
            optionalLabels,
          }}
        >
          {this.props?.children}
        </FormContainerContext.Provider>
      </MountWatcher>
    )
  }
}
