import React, { Suspense, useEffect, useMemo, useState } from 'react'
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
import AdminAPIKeyShowQueryDialog from './AdminAPIKeyShowQueryDialog'

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

function ActionBy(props: {
  label: string
  time?: string
  name?: string
}): React.ReactNode {
  let record: React.ReactNode = 'Never'
  if (props.time && props.name) {
    record = (
      <React.Fragment>
        <Time format='relative' time={props.time} /> by {props.name}
      </React.Fragment>
    )
  } else if (props.time) {
    record = <Time format='relative' time={props.time} />
  }

  return (
    <ListItem divider>
      <ListItemText primary={props.label} secondary={record} />
    </ListItem>
  )
}

export default function AdminAPIKeyDrawer(props: Props): React.JSX.Element {
  const { onClose, apiKeyID } = props
  const classes = useStyles()
  const isOpen = Boolean(apiKeyID)
  const [deleteDialog, setDialogDialog] = useState(false)
  const [editDialog, setEditDialog] = useState(false)
  const [showQuery, setShowQuery] = useState(false)

  // Get API Key triggers/actions
  const context = useMemo(() => ({ additionalTypenames: ['GQLAPIKey'] }), [])
  const [{ data, error }] = useQuery({ query, context })
  const apiKey: GQLAPIKey =
    data?.gqlAPIKeys?.find((d: GQLAPIKey) => {
      return d.id === apiKeyID
    }) || ({} as GQLAPIKey)

  useEffect(() => {
    if (!isOpen) return
    if (!data || apiKey.id) return

    // If the API Key is not found, close the drawer.
    onClose()
  }, [isOpen, data, apiKey.id])

  const lastUsed = apiKey?.lastUsed || null

  if (error) {
    return <GenericError error={error.message} />
  }

  if (isOpen && !apiKey.id) {
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
        <Suspense>
          {showQuery && (
            <AdminAPIKeyShowQueryDialog
              apiKeyID={apiKey.id}
              onClose={() => setShowQuery(false)}
            />
          )}
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
        </Suspense>
        <Grid style={{ width: '30vw' }}>
          <Typography variant='h6' style={{ margin: '16px' }}>
            API Key Details
          </Typography>
          <Divider />
          <List disablePadding>
            <ListItem divider>
              <ListItemText primary='Name' secondary={apiKey.name} />
            </ListItem>
            <ListItem divider>
              <ListItemText
                primary='Description'
                secondary={apiKey.description}
              />
            </ListItem>
            <ListItem divider>
              <ListItemText primary='Role' secondary={apiKey.role} />
            </ListItem>
            <ListItem divider>
              <ListItemText
                primary='Query'
                secondary={
                  <Button
                    variant='outlined'
                    onClick={() => setShowQuery(true)}
                    sx={{ mt: 0.5 }}
                  >
                    Show Query
                  </Button>
                }
              />
            </ListItem>
            <ActionBy
              label='Created'
              time={apiKey.createdAt}
              name={apiKey.createdBy?.name}
            />
            <ActionBy
              label='Updated'
              time={apiKey.updatedAt}
              name={apiKey.updatedBy?.name}
            />
            <ActionBy label='Expires' time={apiKey.expiresAt} />

            <ActionBy
              label='Last Used'
              time={lastUsed?.time}
              name={lastUsed ? lastUsed.ua + ' from ' + lastUsed.ip : ''}
            />
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
