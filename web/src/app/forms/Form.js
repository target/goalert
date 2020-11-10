import React from 'react'
import p from 'prop-types'
import { FormContext } from './context'

/*
 * Form will render a form element and wrap the onSubmit handler
 * to check validation on any nested FormContainers rendered as
 * descendants.
 *
 * onSubmit (if provided) will be called with a second `isValid` argument.
 */
export function Form(props) {
  let _checks = []

  function handleFormSubmit(e) {
    const valid = !_checks.some((f) => !f())
    return props.onSubmit(e, valid)
  }

  function addSubmitCheck(checkFn) {
    _checks.push(checkFn)

    // return function to un-register it
    return () => {
      _checks = _checks.filter((fn) => fn !== checkFn)
    }
  }

  const { onSubmit, disabled, ...formProps } = props

  return (
    <form {...formProps} onSubmit={handleFormSubmit}>
      <FormContext.Provider
        value={{
          disabled,
          addSubmitCheck: addSubmitCheck,
        }}
      >
        {props.children}
      </FormContext.Provider>
    </form>
  )
}

Form.propTypes = {
  onSubmit: p.func,
  disabled: p.bool,
}
