import React from 'react'

export const FormContainerContext = React.createContext({
  onChange: () => {},
  disabled: false,
  errors: [],
  value: {},
  addField: () => () => {},
})
FormContainerContext.displayName = 'FormContainerContext'

export const FormContext = React.createContext({
  disabled: false,
  addSubmitCheck: () => () => {},
})
FormContext.displayName = 'FormContext'
