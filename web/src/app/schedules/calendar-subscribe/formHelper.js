import React from 'react'
import { makeStyles } from '@material-ui/core'
import { CheckCircleOutline as SuccessIcon } from '@material-ui/icons'
import CalenderSuccessForm from './CalendarSuccessForm'

const useStyles = makeStyles(theme => ({
  successIcon: {
    marginRight: theme.spacing(1),
  },
  successTitle: {
    color: 'green',
    display: 'flex',
    alignItems: 'center',
  },
}))

export function FormTitle(isComplete, defaultTitle) {
  const classes = useStyles()

  return isComplete ? (
    <div className={classes.successTitle}>
      <SuccessIcon className={classes.successIcon} />
      Success!
    </div>
  ) : (
    defaultTitle
  )
}

export function getSubtitle(isComplete, defaultSubtitle) {
  const completedSubtitle =
    'Your subscription has been created! You can' +
    ' manage your subscriptions from your profile at any time.'

  return isComplete ? completedSubtitle : defaultSubtitle
}

export function getForm(isComplete, defaultForm, url) {
  return isComplete ? <CalenderSuccessForm url={url} /> : defaultForm
}
