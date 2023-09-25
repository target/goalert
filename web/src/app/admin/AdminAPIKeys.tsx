import React, { useState } from 'react'
import makeStyles from '@mui/styles/makeStyles'
import {
  Button,
  ButtonGroup,
  Grid,
  Typography,
  Card,
  ButtonBase,
} from '@mui/material'
import AdminAPIKeysDrawer from './admin-api-keys/AdminAPIKeysDrawer'
import { GQLAPIKey, CreatedGQLAPIKey } from '../../schema'
import { Time } from '../util/Time'
import { gql, useQuery } from '@apollo/client'
import FlatList, { FlatListListItem } from '../lists/FlatList'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import { Theme } from '@mui/material/styles'
import AdminAPIKeysCreateDialog from './admin-api-keys/AdminAPIKeysCreateDialog'
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
    }
  }
`
const useStyles = makeStyles((theme: Theme) => ({
  root: {
    '& .MuiListItem-root': {
      'border-bottom': '1px solid #333333',
    },
    '& .MuiListItem-root:not(.Mui-selected):hover ': {
      'background-color': '#474747',
    },
  },
  buttons: {
    'padding-bottom': '15px',
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
  const [openCreateAPIKeyDialog, setOpenCreateAPIKeyDialog] = useState(false)
  const [token, setToken] = useState<CreatedGQLAPIKey>({
    id: '',
    token: '',
  })
  const handleOpenCreateDialog = (): void => {
    setOpenCreateAPIKeyDialog(!openCreateAPIKeyDialog)
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
      <div className={classes.root}>
        {selectedAPIKey ? (
          <AdminAPIKeysDrawer
            onClose={() => setSelectedAPIKey(null)}
            apiKey={selectedAPIKey}
          />
        ) : null}
        {openCreateAPIKeyDialog ? (
          <AdminAPIKeysCreateDialog
            onClose={setOpenCreateAPIKeyDialog}
            onTokenDialogClose={onTokenDialogClose}
            setReloadFlag={setReloadFlag}
            setToken={setToken}
          />
        ) : null}
        {tokenDialogClose ? (
          <AdminAPIKeysTokenDialog
            input={token}
            onTokenDialogClose={onTokenDialogClose}
            tokenDialogClose={tokenDialogClose}
          />
        ) : null}
        <Grid item xs={12} container justifyContent='flex-end'>
          <ButtonGroup variant='outlined' className={classes.buttons}>
            <Button data-cy='new' onClick={handleOpenCreateDialog}>
              Create API Key
            </Button>
          </ButtonGroup>
        </Grid>
        <Card
          style={{ width: '100%', paddingLeft: '10px', paddingRight: '10px' }}
          className={
            selectedAPIKey
              ? classes.containerSelected
              : classes.containerDefault
          }
        >
          <FlatList emptyMessage='No Data Available' items={items} />
        </Card>
      </div>
    </React.Fragment>
  )
}
