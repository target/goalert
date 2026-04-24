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
import { SxProps, Theme } from '@mui/material/styles'
import ExpandIcon from '@mui/icons-material/KeyboardArrowDown'
import CollapseIcon from '@mui/icons-material/KeyboardArrowUp'
import toTitleCase from '../util/toTitleCase'
import {
  NoticeType as SchemaNoticeType,
  NotificationStatus,
} from '../../schema'

const classes = {
  gridItem: {
    padding: '4px 0 4px 0',
  },
  lastGridItem: {
    padding: '4px 0 0 0',
  },
  secondGridItem: {
    padding: '8px 0 4px 0',
  },
} satisfies Record<string, SxProps<Theme>>

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
  message: string | JSX.Element
  details?: string | JSX.Element
  endNote?: string | JSX.Element
  action?: React.JSX.Element
}
interface NoticesProps {
  notices?: Notice[]
}

export default function Notices({
  notices = [],
}: NoticesProps): React.JSX.Element | null {
  const [noticesExpanded, setNoticesExpanded] = useState(false)

  if (!notices.length) {
    return null
  }

  function renderShowAllToggle(action?: React.JSX.Element): ReactNode {
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
  function getGridSx(index: number): SxProps<Theme> | undefined {
    switch (index) {
      case 0:
        return undefined
      case 1:
        return classes.secondGridItem
      case notices.length - 1:
        return classes.lastGridItem
      default:
        return classes.gridItem
    }
  }

  function renderNotice(notice: Notice, index: number): React.JSX.Element {
    return (
      <Grid size={12} key={index} sx={getGridSx(index)}>
        <Alert
          severity={toSeverity(notice.type)}
          sx={{
            '& .MuiAlert-message': { width: '100%' },
            '& .MuiAlert-action': { mr: 0 },
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
      <Grid size={12}>
        <Collapse in={noticesExpanded}>
          <Grid container>
            {notices.slice(1).map((n, i) => renderNotice(n, i + 1))}
          </Grid>
        </Collapse>
      </Grid>
    </Grid>
  )
}
