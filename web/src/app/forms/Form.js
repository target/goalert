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
export class Form extends React.PureComponent {
  static propTypes = {
    onSubmit: p.func,
    disabled: p.bool,
  }

  _checks = []

  handleFormSubmit = (e) => {
    const valid = !this._checks.some((f) => !f())
    return this.props.onSubmit(e, valid)
  }

  addSubmitCheck = (checkFn) => {
    this._checks.push(checkFn)

    // return function to un-register it
    return () => {
      this._checks = this._checks.filter((fn) => fn !== checkFn)
    }
  }

  render() {
    const { onSubmit, disabled, ...formProps } = this.props

    return (
      <form {...formProps} onSubmit={this.handleFormSubmit}>
        <FormContext.Provider
          value={{
            disabled,
            addSubmitCheck: this.addSubmitCheck,
          }}
        >
          {this.props.children}
        </FormContext.Provider>
      </form>
    )
  }
}
