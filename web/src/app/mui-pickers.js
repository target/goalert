import React from 'react'
import LocalizationProvider from '@mui/lab/LocalizationProvider'
import DateAdapter from '@mui/lab/AdapterLuxon'

export default function PickersUtilsProvider(props) {
  return (
    <LocalizationProvider dateAdapter={DateAdapter}>
      {props.children}
    </LocalizationProvider>
  )
}
