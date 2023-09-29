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
import AdminAPIKeysDrawer from './admin-api-keys/AdminAPIKeysDrawer'
import { GQLAPIKey, CreatedGQLAPIKey } from '../../schema'
import { Time } from '../util/Time'
import { gql, useQuery } from '@apollo/client'
import FlatList, { FlatListListItem } from '../lists/FlatList'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import { Theme } from '@mui/material/styles'
import AdminAPIKeysActionDialog from './admin-api-keys/AdminAPIKeysActionDialog'
import AdminAPIKeysTokenDialog from './admin-api-keys/AdminAPIKeysTokenDialog'

const getAPIKeysQuery = gql`
  query gqlAPIKeysQuery {
    gqlAPIKeys {
      id
      name
      description
      createdAt
      createdBy {
        id
        role
        name
        email
      }
      updatedAt
      updatedBy {
        id
        role
        name
        email
      }
      lastUsed {
        time
        ua
        ip
      }
      expiresAt
      allowedFields
      role
    }
  }
`
const useStyles = makeStyles((theme: Theme) => ({
  root: {
    '& .MuiListItem-root': {
      'border-bottom': '1px solid #333333',
    },
  },
  buttons: {
    'margin-bottom': '15px',
  },
  containerDefault: {
    [theme.breakpoints.up('md')]: {
      maxWidth: '100%',
      transition: `max-width ${theme.transitions.duration.leavingScreen}ms ease`,
    },
    '& .MuiListItem-root': {
      padding: '0px',
    },
  },
  containerSelected: {
    [theme.breakpoints.up('md')]: {
      maxWidth: '70%',
      transition: `max-width ${theme.transitions.duration.enteringScreen}ms ease`,
    },
    '& .MuiListItem-root': {
      padding: '0px',
    },
  },
}))

export default function AdminAPIKeys(): JSX.Element {
  const classes = useStyles()
  const [selectedAPIKey, setSelectedAPIKey] = useState<GQLAPIKey | null>(null)
  const [reloadFlag, setReloadFlag] = useState<number>(0)
  const [tokenDialogClose, onTokenDialogClose] = useState(false)
  const [openActionAPIKeyDialog, setOpenActionAPIKeyDialog] = useState(false)
  const [create, setCreate] = useState(false)
  const [apiKey, setAPIKey] = useState<GQLAPIKey>({
    id: '',
    name: '',
    description: '',
    createdAt: '',
    createdBy: null,
    updatedAt: '',
    updatedBy: null,
    lastUsed: null,
    expiresAt: '',
    allowedFields: [],
    role: 'user',
  })
  const [token, setToken] = useState<CreatedGQLAPIKey>({
    id: '',
    token: '',
  })
  const handleOpenCreateDialog = (): void => {
    setCreate(true)
    setOpenActionAPIKeyDialog(!openActionAPIKeyDialog)
  }
  const { data, loading, error } = useQuery(getAPIKeysQuery, {
    variables: {
      reloadData: reloadFlag,
    },
  })

  if (error) {
    return <GenericError error={error.message} />
  }

  if (loading && !data) {
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
                {key.allowedFields.length + ' allowed fields (read-only)'}
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
            if (!openActionAPIKeyDialog) {
              setSelectedAPIKey(null)
            }
          }}
          apiKey={selectedAPIKey}
          setCreate={setCreate}
          setOpenActionAPIKeyDialog={setOpenActionAPIKeyDialog}
          setAPIKey={setAPIKey}
        />
      ) : null}
      {openActionAPIKeyDialog ? (
        <AdminAPIKeysActionDialog
          onClose={() => {
            if (!create && selectedAPIKey) {
              selectedAPIKey.name = apiKey.name
              selectedAPIKey.description = apiKey.description
              setSelectedAPIKey(selectedAPIKey)
            }

            setOpenActionAPIKeyDialog(false)
          }}
          onTokenDialogClose={onTokenDialogClose}
          setReloadFlag={setReloadFlag}
          setToken={setToken}
          create={create}
          apiKey={apiKey}
          setAPIKey={setAPIKey}
          setSelectedAPIKey={setSelectedAPIKey}
        />
      ) : null}
      {tokenDialogClose ? (
        <AdminAPIKeysTokenDialog
          input={token}
          onTokenDialogClose={onTokenDialogClose}
          tokenDialogClose={tokenDialogClose}
        />
      ) : null}
      <Card
        style={{ width: '100%', padding: '10px' }}
        className={
          selectedAPIKey ? classes.containerSelected : classes.containerDefault
        }
      >
        <CardHeader
          title='API Key List'
          component='h2'
          sx={{ paddingBottom: 0, margin: 0 }}
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
