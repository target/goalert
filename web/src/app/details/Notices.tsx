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
  gridItem: {
    padding: '4px 0 4px 0',
  },
  lastGridItem: {
    padding: '4px 0 0 0',
  },
  secondGridItem: {
    padding: '8px 0 4px 0',
  },
})

interface NoticesProps {
  notices: Notice[] // checks against .length are safe as notices is required
}

export default function Notices(props: NoticesProps): JSX.Element {
  const classes = useStyles()
  const [alertsExpanded, setAlertsExpanded] = useState(false)

  function renderShowAllToggle(): ReactNode {
    if (props.notices.length <= 1) return null

    return (
      <Badge
        color='primary'
        badgeContent={props.notices.length - 1}
        invisible={alertsExpanded}
      >
        <IconButton onClick={() => setAlertsExpanded(!alertsExpanded)}>
          {alertsExpanded ? <CollapseIcon /> : <ExpandIcon />}
        </IconButton>
      </Badge>
    )
  }

  /*
   * Spacing set manually on grid items to accomadate manual
   * accordian transitions for multiple notices
   */
  function getGridClassName(index: number): string {
    switch (index) {
      case 0:
        return ''
      case 1:
        return classes.secondGridItem
      case props.notices.length - 1:
        return classes.lastGridItem
      default:
        return classes.gridItem
    }
  }

  function renderAlert(notice: Notice, index: number): JSX.Element {
    return (
      <Grid key={index} className={getGridClassName(index)} item xs={12}>
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
    <Grid container>
      {renderAlert(props.notices[0], 0)}
      <Grid item xs={12}>
        <Collapse in={alertsExpanded}>
          <Grid container>
            {props.notices.slice(1).map((n, i) => renderAlert(n, i + 1))}
          </Grid>
        </Collapse>
      </Grid>
    </Grid>
  )
}
