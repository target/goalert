import React, { useRef } from 'react'
import { FormContext } from './context'

type FormProps = {
  onSubmit: any,
  disabled: boolean,
  // children: Node,
  className: string,
}
/*
 * Form will render a form element and wrap the onSubmit handler
 * to check validation on any nested FormContainers rendered as
 * descendants.
 *
 * onSubmit (if provided) will be called with a second `isValid` argument.
 */
export function Form(props: FormProps): JSX.Element {
  const checks = useRef([])

  function handleFormSubmit(e: any) {
    const valid = !checks.current.some((f: any) => !f())
    return props.onSubmit(e, valid)
  }

  function addSubmitCheck(checkFn: never) {
    checks.current.push(checkFn)

    // return function to un-register it
    return () => {
      checks.current = checks.current.filter((fn: string) => fn !== checkFn)
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
