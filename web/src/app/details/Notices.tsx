import React from 'react'
import { makeStyles } from '@material-ui/core'
import { Alert, AlertTitle, AlertProps } from '@material-ui/lab'
import { toTitleCase } from '../util/toTitleCase'
import { Notice } from '../../schema'

const useStyles = makeStyles({
  alertMessage: {
    width: '100%',
  },
})

interface NoticesProps {
  notices: Notice[]
}

export default function Notices(props: NoticesProps) {
  const classes = useStyles()

  function renderAlert(notice: Notice, index: number) {
    return (
      <Alert
        key={index}
        severity={notice.type.toLowerCase() as AlertProps['severity']}
        classes={{ message: classes.alertMessage }}
      >
        <AlertTitle>
          {toTitleCase(notice.type)}: {notice.message}
        </AlertTitle>
        {notice.details}
      </Alert>
    )
  }

  return props.notices.map((n, i) => renderAlert(n, i))
}
