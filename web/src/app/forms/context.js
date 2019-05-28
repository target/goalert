import React from 'react'

export const FormContainerContext = React.createContext({
  onChange: (field, value) => {},
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
