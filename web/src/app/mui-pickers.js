import React from 'react'
import AdapterLuxon from '@mui/lab/AdapterLuxon'
import LocalizationProvider from '@mui/lab/LocalizationProvider'

export default function PickersUtilsProvider(props) {
  return (
    <LocalizationProvider dateAdapter={AdapterLuxon}>
      {props.children}
    </LocalizationProvider>
  )
}
