import React, { useState } from 'react'
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
  Button,
  ButtonGroup,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { GQLAPIKey } from '../../../schema'
import AdminAPIKeyDeleteDialog from './AdminAPIKeyDeleteDialog'
import AdminAPIKeyEditDialog from './AdminAPIKeyEditDialog'
import { Time } from '../../util/Time'
import { gql, useQuery } from 'urql'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'

// query for getting existing API Keys
const query = gql`
  query gqlAPIKeysQuery {
    gqlAPIKeys {
      id
      name
      description
      createdAt
      createdBy {
        id
        name
      }
      updatedAt
      updatedBy {
        id
        name
      }
      lastUsed {
        time
        ua
        ip
      }
      expiresAt
      query
      role
    }
  }
`

// property for this object
interface Props {
  onClose: () => void
  apiKeyID?: string
  onDuplicateClick: () => void
}

const useStyles = makeStyles(() => ({
  buttons: {
    textAlign: 'right',
    width: '30vw',
    padding: '15px 10px',
  },
}))

export default function AdminAPIKeyDrawer(props: Props): JSX.Element {
  const { onClose, apiKeyID } = props
  const classes = useStyles()
  const isOpen = Boolean(apiKeyID)
  const [deleteDialog, setDialogDialog] = useState(false)
  const [editDialog, setEditDialog] = useState(false)

  // Get API Key triggers/actions
  const [{ data, fetching, error }] = useQuery({ query })
  const apiKey: GQLAPIKey =
    data?.gqlAPIKeys?.find((d: GQLAPIKey) => {
      return d.id === apiKeyID
    }) || ({} as GQLAPIKey)

  if (error) {
    return <GenericError error={error.message} />
  }

  if (fetching && !data) {
    return <Spinner />
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
        {deleteDialog ? (
          <AdminAPIKeyDeleteDialog
            onClose={(yes: boolean): void => {
              setDialogDialog(false)

              if (yes) {
                onClose()
              }
            }}
            apiKeyID={apiKey.id}
          />
        ) : null}
        {editDialog ? (
          <AdminAPIKeyEditDialog
            onClose={() => setEditDialog(false)}
            apiKeyID={apiKey.id}
          />
        ) : null}
        <Grid style={{ width: '30vw' }}>
          <Typography variant='h6' style={{ margin: '16px' }}>
            API Key Details
          </Typography>
          <Divider />
          <List disablePadding>
            <ListItem divider>
              <ListItemText primary='Name' secondary={apiKey?.name} />
            </ListItem>
            <ListItem divider>
              <ListItemText
                primary='Description'
                secondary={apiKey?.description}
              />
            </ListItem>
            <ListItem divider>
              <ListItemText
                primary='Query'
                secondary={<Button variant='text'>Show Query</Button>}
              />
            </ListItem>
            <ListItem divider>
              <ListItemText
                primary='Creation Time'
                secondary={<Time prefix='' time={apiKey?.createdAt} />}
              />
            </ListItem>
            <ListItem divider>
              <ListItemText
                primary='Created By'
                secondary={apiKey?.createdBy?.name}
              />
            </ListItem>
            <ListItem divider>
              <ListItemText
                primary='Expires At'
                secondary={<Time prefix='' time={apiKey?.expiresAt} />}
              />
            </ListItem>
            <ListItem divider>
              <ListItemText
                primary='Updated By'
                secondary={apiKey?.updatedBy?.name}
              />
            </ListItem>
            <ListItem divider>
              <ListItemText primary='Role' secondary={apiKey?.role} />
            </ListItem>
          </List>
          <Grid className={classes.buttons}>
            <ButtonGroup variant='contained'>
              <Button onClick={() => setDialogDialog(true)}>Delete</Button>
              <Button onClick={() => setEditDialog(true)}>Edit</Button>
              <Button
                onClick={() => props.onDuplicateClick()}
                title='Create a new API Key with the same settings as this one.'
              >
                Duplicate
              </Button>
            </ButtonGroup>
          </Grid>
        </Grid>
      </Drawer>
    </ClickAwayListener>
  )
}
