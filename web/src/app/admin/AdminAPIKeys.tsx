import React, { useState } from 'react'
import makeStyles from '@mui/styles/makeStyles'
import {
  Button,
  Grid,
  Typography,
  Card,
  ButtonBase,
  CardHeader,
} from '@mui/material'
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
// query for getting existing API Keys
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
      subText: (
        <ButtonBase
          onClick={() => {
            setSelectedAPIKey(key)
          }}
          style={{ width: '100%', textAlign: 'left', padding: '5px 15px' }}
        >
          <Grid container>
            <Grid item xs justifyContent='flex-start'>
              <Typography gutterBottom variant='subtitle1' component='div'>
                {key.name}
              </Typography>
              <Typography gutterBottom variant='subtitle2' component='div'>
                <Time prefix='Expires At: ' time={key.expiresAt} />
              </Typography>
              <Typography gutterBottom variant='subtitle2' component='div'>
                {key.allowedFields.length +
                  ' allowed fields' +
                  (key.allowedFields.some((f) => f.startsWith('Mutation.'))
                    ? ''
                    : ' (read-only)')}
              </Typography>
            </Grid>
            <Grid item>
              <Typography gutterBottom variant='subtitle2' component='div'>
                <Time prefix='Last Used: ' time={key.expiresAt} />
              </Typography>
            </Grid>
          </Grid>
        </ButtonBase>
      ),
    }),
  )

  return (
    <React.Fragment>
      {selectedAPIKey ? (
        <AdminAPIKeysDrawer
          onClose={() => {
            setSelectedAPIKey(null)
          }}
          apiKeyId={selectedAPIKey.id}
        />
      ) : null}
      {createAPIKeyDialogClose ? (
        <AdminAPIKeyCreateDialog
          onClose={() => {
            onCreateAPIKeyDialogClose(false)
          }}
        />
      ) : null}
      <Card
        sx={{ width: '100%', padding: '0px' }}
        className={
          selectedAPIKey ? classes.containerSelected : classes.containerDefault
        }
      >
        <CardHeader
          title='API Key List'
          component='h2'
          sx={{ padding: '20px 25px 0px', margin: 0 }}
          action={
            <Button
              data-cy='new'
              variant='contained'
              className={classes.buttons}
              onClick={handleOpenCreateDialog}
              startIcon={<Add />}
            >
              Create API Key
            </Button>
          }
        />
        <FlatList emptyMessage='No Data Available' items={items} />
      </Card>
    </React.Fragment>
  )
}
