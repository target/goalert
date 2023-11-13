import React, { useState, ReactNode } from 'react'
import {
  Badge,
  Collapse,
  Grid,
  IconButton,
  Alert,
  AlertTitle,
  AlertColor,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import ExpandIcon from '@mui/icons-material/KeyboardArrowDown'
import CollapseIcon from '@mui/icons-material/KeyboardArrowUp'
import toTitleCase from '../util/toTitleCase'
import {
  NoticeType as SchemaNoticeType,
  NotificationStatus,
} from '../../schema'

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

export type NoticeType = SchemaNoticeType | AlertColor | NotificationStatus
function assertNever(x: never): never {
  throw new Error('Unexpected value: ' + x)
}
export function toSeverity(notice: NoticeType): AlertColor {
  switch (notice) {
    case 'success':
    case 'OK':
      return 'success'
    case 'warning':
    case 'WARNING':
    case 'WARN':
      return 'warning'
    case 'error':
    case 'ERROR':
      return 'error'
    case 'info':
    case 'INFO':
      return 'info'
    default:
      assertNever(notice)
  }
}

export interface Notice {
  type: NoticeType
  message: string | React.ReactNode
  details?: string | React.ReactNode
  endNote?: string | React.ReactNode
  action?: React.ReactNode
}
interface NoticesProps {
  notices?: Notice[]
}

export default function Notices({
  notices = [],
}: NoticesProps): React.ReactNode | null {
  const classes = useStyles()
  const [noticesExpanded, setNoticesExpanded] = useState(false)

  if (!notices.length) {
    return null
  }

  function renderShowAllToggle(action?: React.ReactNode): ReactNode {
    if (notices.length <= 1) return null
    return (
      <React.Fragment>
        {action}
        <Badge
          color='primary'
          badgeContent={notices.length - 1}
          invisible={noticesExpanded}
        >
          <IconButton
            onClick={() => setNoticesExpanded(!noticesExpanded)}
            size='large'
            sx={{ pl: 1 }}
          >
            {noticesExpanded ? <CollapseIcon /> : <ExpandIcon />}
          </IconButton>
        </Badge>
      </React.Fragment>
    )
  }

  /*
   * Spacing set manually on grid items to accommodate manual
   * accordion transitions for multiple notices
   */
  function getGridClassName(index: number): string {
    switch (index) {
      case 0:
        return ''
      case 1:
        return classes.secondGridItem
      case notices.length - 1:
        return classes.lastGridItem
      default:
        return classes.gridItem
    }
  }

  function renderNotice(notice: Notice, index: number): React.ReactNode {
    return (
      <Grid key={index} className={getGridClassName(index)} item xs={12}>
        <Alert
          severity={toSeverity(notice.type)}
          classes={{
            message: classes.alertMessage,
            action: classes.alertAction,
          }}
          elevation={1}
          action={
            <div
              style={{ display: 'flex', alignItems: 'center', height: '100%' }}
            >
              {index === 0 && notices.length > 1
                ? renderShowAllToggle(notice.action)
                : notice.action}
            </div>
          }
        >
          <AlertTitle>
            {toTitleCase(notice.type)}: {notice.message}
          </AlertTitle>
          {notice.details}
          {notice.endNote && (
            <div style={{ float: 'right' }}>{notice.endNote}</div>
          )}
        </Alert>
      </Grid>
    )
  }

  return (
    <Grid container>
      {renderNotice(notices[0], 0)}
      <Grid item xs={12}>
        <Collapse in={noticesExpanded}>
          <Grid container>
            {notices.slice(1).map((n, i) => renderNotice(n, i + 1))}
          </Grid>
        </Collapse>
      </Grid>
    </Grid>
  )
}
