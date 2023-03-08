import React from 'react'
import { LocalizationProvider } from '@mui/x-date-pickers'
import { AdapterLuxon } from '@mui/x-date-pickers/AdapterLuxon'

export default function PickersUtilsProvider(props) {
  return (
    <LocalizationProvider dateAdapter={AdapterLuxon}>
      {props.children}
    </LocalizationProvider>
  )
}
