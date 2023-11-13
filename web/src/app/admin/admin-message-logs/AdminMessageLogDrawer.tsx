import React from 'react'
import {
  ClickAwayListener,
  Divider,
  Drawer,
  Grid,
  List,
  ListItem,
  ListItemText,
  Toolbar,
  Typography,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import { OpenInNew } from '@mui/icons-material'
import { DateTime } from 'luxon'
import AppLink from '../../util/AppLink'
import { DebugMessage } from '../../../schema'

interface Props {
  onClose: () => void
  log: DebugMessage | null
}

const useStyles = makeStyles((theme: Theme) => ({
  appLink: {
    display: 'flex',
    alignItems: 'center',
  },
  appLinkIcon: {
    paddingLeft: theme.spacing(0.5),
  },
}))

export default function AdminMessageLogDrawer(props: Props): React.ReactNode {
  const { onClose, log } = props
  const classes = useStyles()

  const isOpen = Boolean(log)

  const sentAtText = (): string => {
    return log?.sentAt
      ? DateTime.fromISO(log.sentAt).toFormat('fff')
      : 'Not Sent'
  }

  return (
    <ClickAwayListener onClickAway={onClose} mouseEvent='onMouseUp'>
      <Drawer
        anchor='right'
        open={isOpen}
        variant='persistent'
        data-cy='debug-message-details'
      >
        <Toolbar />
        <Grid style={{ width: '30vw' }}>
          <Typography variant='h6' style={{ margin: '16px' }}>
            Log Details
          </Typography>
          <Divider />
          <List disablePadding>
            {!!log?.id && (
              <ListItem divider>
                <ListItemText primary='ID' secondary={log.id} />
              </ListItem>
            )}
            {!!log?.createdAt && (
              <ListItem divider>
                <ListItemText
                  primary='Created At'
                  secondary={DateTime.fromISO(log.createdAt).toFormat('fff')}
                />
              </ListItem>
            )}
            {!!log?.updatedAt && (
              <ListItem divider>
                <ListItemText
                  primary='Updated At'
                  secondary={DateTime.fromISO(log.updatedAt).toFormat('fff')}
                />
              </ListItem>
            )}
            <ListItem divider>
              <ListItemText primary='Sent At' secondary={sentAtText()} />
            </ListItem>
            {!!log?.retryCount && (
              <ListItem divider>
                <ListItemText
                  primary='Retry Count'
                  secondary={log.retryCount}
                />
              </ListItem>
            )}

            {!!log?.type && (
              <ListItem divider>
                <ListItemText
                  primary='Notification Type'
                  secondary={log.type}
                />
              </ListItem>
            )}
            {!!log?.status && (
              <ListItem divider>
                <ListItemText primary='Current Status' secondary={log.status} />
              </ListItem>
            )}

            {!!log?.userID && !!log?.userName && (
              <ListItem divider>
                <ListItemText
                  primary='User'
                  secondary={
                    <AppLink
                      className={classes.appLink}
                      to={`/users/${log?.userID}`}
                      newTab
                    >
                      {log.userName}
                      <OpenInNew
                        className={classes.appLinkIcon}
                        fontSize='small'
                      />
                    </AppLink>
                  }
                  secondaryTypographyProps={{ component: 'div' }}
                />
              </ListItem>
            )}
            {!!log?.serviceID && !!log?.serviceName && (
              <ListItem divider>
                <ListItemText
                  primary='Service'
                  secondary={
                    <AppLink
                      className={classes.appLink}
                      to={`/services/${log.serviceID}`}
                      newTab
                    >
                      {log.serviceName}
                      <OpenInNew
                        className={classes.appLinkIcon}
                        fontSize='small'
                      />
                    </AppLink>
                  }
                  secondaryTypographyProps={{ component: 'div' }}
                />
              </ListItem>
            )}
            {!!log?.alertID && (
              <ListItem divider>
                <ListItemText
                  primary='Alert'
                  secondary={
                    <AppLink
                      className={classes.appLink}
                      to={`/alerts/${log.alertID}`}
                      newTab
                    >
                      {log.alertID}
                      <OpenInNew
                        className={classes.appLinkIcon}
                        fontSize='small'
                      />
                    </AppLink>
                  }
                  secondaryTypographyProps={{ component: 'div' }}
                />
              </ListItem>
            )}

            {!!log?.source && (
              <ListItem divider>
                <ListItemText primary='Source' secondary={log.source} />
              </ListItem>
            )}
            {!!log?.destination && (
              <ListItem divider>
                <ListItemText
                  primary='Destination'
                  secondary={log.destination}
                />
              </ListItem>
            )}
            {!!log?.providerID && (
              <ListItem divider>
                <ListItemText
                  primary='Provider ID'
                  secondary={log.providerID}
                />
              </ListItem>
            )}
          </List>
        </Grid>
      </Drawer>
    </ClickAwayListener>
  )
}
