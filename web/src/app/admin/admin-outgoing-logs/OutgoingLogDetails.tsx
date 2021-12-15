import React from 'react'
import Drawer from '@material-ui/core/Drawer'
import Grid from '@material-ui/core/Grid'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Toolbar from '@material-ui/core/Toolbar'
import { DebugMessage } from './AdminOutgoingLogs'
import AppLink from '../../util/AppLink'

interface Props {
  open: boolean
  onClose: () => void
  log?: DebugMessage | null
}

export default function OutgoingLogDetails(props: Props): JSX.Element {
  const { open, onClose, log } = props

  return (
    <Drawer anchor='right' open={open} variant='persistent' onClose={onClose}>
      <Toolbar />
      <Grid style={{ width: '30vw' }}>
        <List>
          {log?.id && (
            <ListItem divider>
              <ListItemText primary='ID' secondary={log?.id} />
            </ListItem>
          )}
          {log?.createdAt && (
            <ListItem divider>
              <ListItemText primary='Created At' secondary={log?.createdAt} />
            </ListItem>
          )}
          {log?.updatedAt && (
            <ListItem divider>
              <ListItemText primary='Updated At' secondary={log?.updatedAt} />
            </ListItem>
          )}
          {log?.type && (
            <ListItem divider>
              <ListItemText primary='Notification Type' secondary={log?.type} />
            </ListItem>
          )}
          {log?.status && (
            <ListItem divider>
              <ListItemText primary='Current Status' secondary={log?.status} />
            </ListItem>
          )}

          {/* TODO: use AppLink to point to user/service/alert pages */}
          {log?.userID && log?.userName && (
            <ListItem divider>
              <ListItemText
                primary='User'
                secondary={
                  <AppLink to={`/users/${log?.userID}`} newTab icon>
                    {log?.userName}
                  </AppLink>
                }
                secondaryTypographyProps={{ component: 'div' }}
              />
            </ListItem>
          )}
          {log?.serviceID && log?.serviceName && (
            <ListItem divider>
              <ListItemText
                primary='Service'
                secondary={
                  <AppLink to={`/services/${log?.serviceID}`} newTab icon>
                    {log?.serviceName}
                  </AppLink>
                }
                secondaryTypographyProps={{ component: 'div' }}
              />
            </ListItem>
          )}
          {log?.alertID && (
            <ListItem divider>
              <ListItemText
                primary='Alert'
                secondary={
                  <AppLink to={`/alerts/${log?.alertID}`} newTab icon>
                    {log?.serviceName}
                  </AppLink>
                }
                secondaryTypographyProps={{ component: 'div' }}
              />
            </ListItem>
          )}

          {log?.source && (
            <ListItem divider>
              <ListItemText primary='Source' secondary={log?.source} />
            </ListItem>
          )}
          {log?.destination && (
            <ListItem divider>
              <ListItemText
                primary='Destination'
                secondary={log?.destination}
              />
            </ListItem>
          )}
          {log?.providerID && (
            <ListItem divider>
              <ListItemText primary='Provider ID' secondary={log?.providerID} />
            </ListItem>
          )}
        </List>
      </Grid>
    </Drawer>
  )
}
