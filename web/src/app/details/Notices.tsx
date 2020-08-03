import React, { useState, ReactNode } from 'react'
import {
  Badge,
  Collapse,
  Grid,
  IconButton,
  makeStyles,
} from '@material-ui/core'
import ExpandIcon from '@material-ui/icons/KeyboardArrowDown'
import CollapseIcon from '@material-ui/icons/KeyboardArrowUp'
import { Alert, AlertTitle, AlertProps } from '@material-ui/lab'
import { toTitleCase } from '../util/toTitleCase'
import { Notice } from '../../schema'

const useStyles = makeStyles({
  alertAction: {
    marginRight: 0,
  },
  alertMessage: {
    width: '100%',
  },
})

interface NoticesProps {
  notices: Notice[] // checks against .length are safe as notices is required
}

export default function Notices(props: NoticesProps) {
  const classes = useStyles()
  const [alertsExpanded, setAlertsExpanded] = useState(false)

  function renderShowAllToggle(): ReactNode {
    if (props.notices.length <= 1) return null

    if (alertsExpanded) {
      return (
        <IconButton onClick={() => setAlertsExpanded(false)}>
          <CollapseIcon />
        </IconButton>
      )
    }

    return (
      <Badge color='primary' badgeContent={props.notices.length - 1}>
        <IconButton onClick={() => setAlertsExpanded(true)}>
          <ExpandIcon />
        </IconButton>
      </Badge>
    )
  }

  function renderAlert(notice: Notice, index: number) {
    return (
      <Grid key={index} item xs={12}>
        <Alert
          severity={notice.type.toLowerCase() as AlertProps['severity']}
          classes={{
            message: classes.alertMessage,
            action: classes.alertAction,
          }}
          elevation={1}
          action={index === 0 ? renderShowAllToggle() : null}
        >
          <AlertTitle>
            {toTitleCase(notice.type)}: {notice.message}
          </AlertTitle>
          {notice.details}
        </Alert>
      </Grid>
    )
  }

  return (
    <Grid container spacing={1}>
      {renderAlert(props.notices[0], 0)}
      <Grid item xs={12}>
        <Collapse in={alertsExpanded}>
          <Grid container spacing={1}>
            {props.notices.slice(1).map((n, i) => renderAlert(n, i + 1))}
          </Grid>
        </Collapse>
      </Grid>
    </Grid>
  )
}
