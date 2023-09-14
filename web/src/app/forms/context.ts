import React from 'react'

// interface FormContainerContextProps {
//   Provider: any;
//   displayName: string;
//   onChange: any; 
//   disabled: boolean; 
//   errors: never[] | any[]; 
//   value: {}; 
//   addField: any; 
//   optionalLabels: boolean; 
// }

export const FormContainerContext = React.createContext({
  onChange: () => {},
  disabled: false,
  errors: [],
  value: {},
  addField: (fieldName: string, validate: any) => () => {},
  optionalLabels: false,
})
FormContainerContext.displayName = 'FormContainerContext'

export const FormContext = React.createContext({
  disabled: false,
  addSubmitCheck: (checkFn: never) => () => {},
})
FormContext.displayName = 'FormContext'
