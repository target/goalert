import React, { useRef } from 'react'
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
  const checks = useRef([])

  function handleFormSubmit(e) {
    const valid = !checks.current.some((f) => !f())
    return props.onSubmit(e, valid)
  }

  function addSubmitCheck(checkFn) {
    checks.current.push(checkFn)

    // return function to un-register it
    return () => {
      checks.current = checks.current.filter((fn) => fn !== checkFn)
    }
  }

  const { onSubmit, disabled, ...formProps } = props

  return (
    <form {...formProps} onSubmit={handleFormSubmit}>
      <FormContext.Provider
        value={{
          disabled,
          addSubmitCheck,
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
  children: p.node,
  className: p.string,
}