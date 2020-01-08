import React from 'react'
import { makeStyles } from '@material-ui/core'
import { CheckCircleOutline as SuccessIcon } from '@material-ui/icons'
import CalenderSuccessForm from './CalendarSuccessForm'

const useStyles = makeStyles({
  successTitle: {
    color: 'green',
    display: 'flex',
    alignItems: 'center',
  },
})

export function FormTitle(isComplete, defaultTitle) {
  const classes = useStyles()

  return isComplete ? (
    <div className={classes.successTitle}>
      <SuccessIcon />
      &nbsp;Success!
    </div>
  ) : (
    defaultTitle
  )
}

export function getForm(isComplete, defaultForm, url) {
  return isComplete ? <CalenderSuccessForm url={url} /> : defaultForm
}
