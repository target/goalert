import React, { useState } from 'react'
import makeStyles from '@mui/styles/makeStyles'
import { Button, Grid, Typography, Card } from '@mui/material'
import { Add } from '@mui/icons-material'
import AdminAPIKeysDrawer from './admin-api-keys/AdminAPIKeyDrawer'
import { GQLAPIKey } from '../../schema'
import { Time } from '../util/Time'
import { gql, useQuery } from 'urql'
import FlatList, { FlatListListItem } from '../lists/FlatList'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import { Theme } from '@mui/material/styles'
import AdminAPIKeyCreateDialog from './admin-api-keys/AdminAPIKeyCreateDialog'
import AdminAPIKeyDeleteDialog from './admin-api-keys/AdminAPIKeyDeleteDialog'
import AdminAPIKeyEditDialog from './admin-api-keys/AdminAPIKeyEditDialog'
import OtherActions from '../util/OtherActions'

const query = gql`
  query gqlAPIKeysQuery {
    gqlAPIKeys {
      id
      name
      lastUsed {
        time
        ua
        ip
      }
      expiresAt
      allowedFields
    }
  }
`

const useStyles = makeStyles((theme: Theme) => ({
  buttons: {
    'margin-bottom': '15px',
  },
  containerDefault: {
    [theme.breakpoints.up('md')]: {
      maxWidth: '100%',
      transition: `max-width ${theme.transitions.duration.leavingScreen}ms ease`,
    },
  },
  containerSelected: {
    [theme.breakpoints.up('md')]: {
      maxWidth: '70%',
      transition: `max-width ${theme.transitions.duration.enteringScreen}ms ease`,
    },
  },
}))

export default function AdminAPIKeys(): JSX.Element {
  const classes = useStyles()
  const [selectedAPIKey, setSelectedAPIKey] = useState<GQLAPIKey | null>(null)
  const [createAPIKeyDialogClose, onCreateAPIKeyDialogClose] = useState(false)
  const [editDialog, setEditDialog] = useState<string | undefined>()
  const [deleteDialog, setDeleteDialog] = useState<string | undefined>()

  // handles the openning of the create dialog form which is used for creating new API Key
  const handleOpenCreateDialog = (): void => {
    onCreateAPIKeyDialogClose(!createAPIKeyDialogClose)
  }

  // Get API Key triggers/actions
  const [{ data, fetching, error }] = useQuery({ query })

  if (error) {
    return <GenericError error={error.message} />
  }

  if (fetching && !data) {
    return <Spinner />
  }

  const items = data.gqlAPIKeys.map(
    (key: GQLAPIKey): FlatListListItem => ({
      selected: (key as GQLAPIKey).id === selectedAPIKey?.id,
      highlight: (key as GQLAPIKey).id === selectedAPIKey?.id,
      primaryText: <Typography>{key.name}</Typography>,
      disableTypography: true,
      subText: (
        <React.Fragment>
          <Typography variant='subtitle2' component='div' color='textSecondary'>
            <Time prefix='Expires At: ' time={key.expiresAt} />
          </Typography>
          <Typography variant='subtitle2' component='div' color='textSecondary'>
            {key.allowedFields.length +
              ' allowed fields' +
              (key.allowedFields.some((f) => f.startsWith('Mutation.'))
                ? ''
                : ' (read-only)')}
          </Typography>
        </React.Fragment>
      ),
      secondaryAction: (
        <Grid container>
          <Grid
            item
            xs={12}
            sx={{ display: 'flex', justifyContent: 'flex-end' }}
          >
            <Typography
              gutterBottom
              variant='subtitle2'
              component='div'
              color='textSecondary'
            >
              <Time prefix='Last Used: ' time={key.expiresAt} />
            </Typography>
          </Grid>
          <Grid
            item
            xs={12}
            sx={{ display: 'flex', justifyContent: 'flex-end' }}
          >
            <OtherActions
              actions={[
                {
                  label: 'Edit',
                  onClick: () => setEditDialog(key.id),
                },
                {
                  label: 'Delete',
                  onClick: () => setDeleteDialog(key.id),
                },
              ]}
            />
          </Grid>
        </Grid>
      ),
      onClick: () => setSelectedAPIKey(key),
    }),
  )

  return (
    <React.Fragment>
      <AdminAPIKeysDrawer
        onClose={() => {
          setSelectedAPIKey(null)
        }}
        apiKeyID={selectedAPIKey?.id}
      />
      {createAPIKeyDialogClose ? (
        <AdminAPIKeyCreateDialog
          onClose={() => {
            onCreateAPIKeyDialogClose(false)
          }}
        />
      ) : null}
      {deleteDialog ? (
        <AdminAPIKeyDeleteDialog
          onClose={(): void => {
            setDeleteDialog('')
          }}
          apiKeyID={deleteDialog}
        />
      ) : null}
      {editDialog ? (
        <AdminAPIKeyEditDialog
          onClose={() => setEditDialog('')}
          apiKeyID={editDialog}
        />
      ) : null}
      <div
        className={
          selectedAPIKey ? classes.containerSelected : classes.containerDefault
        }
      >
        <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
          <Button
            data-cy='new'
            variant='contained'
            className={classes.buttons}
            onClick={handleOpenCreateDialog}
            startIcon={<Add />}
          >
            Create API Key
          </Button>
        </div>
        <Card sx={{ width: '100%', padding: '0px' }}>
          <FlatList emptyMessage='No Results' items={items} />
        </Card>
      </div>
    </React.Fragment>
  )
}
