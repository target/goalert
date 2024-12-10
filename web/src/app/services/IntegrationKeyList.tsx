import React, { Suspense, useState } from 'react'
import { gql, useQuery } from 'urql'
import Button from '@mui/material/Button'
import Grid from '@mui/material/Grid'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CreateFAB from '../lists/CreateFAB'
import FlatList, { FlatListListItem } from '../lists/FlatList'
import IconButton from '@mui/material/IconButton'
import { Trash } from '../icons'
import IntegrationKeyCreateDialog from './IntegrationKeyCreateDialog'
import IntegrationKeyDeleteDialog from './IntegrationKeyDeleteDialog'
import CopyText from '../util/CopyText'
import AppLink from '../util/AppLink'
import { useIsWidthDown } from '../util/useWidth'
import { Add } from '@mui/icons-material'
import makeStyles from '@mui/styles/makeStyles'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import { IntegrationKey } from '../../schema'
import { useFeatures } from '../util/RequireConfig'
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Chip,
  Typography,
} from '@mui/material'
import { ChevronDown } from 'mdi-material-ui'

const query = gql`
  query ($serviceID: ID!) {
    service(id: $serviceID) {
      id # need to tie the result to the correct record
      integrationKeys {
        id
        type
        name
        href
        externalSystemName
      }
    }
  }
`

const useStyles = makeStyles({
  copyIcon: {
    paddingRight: '0.25em',
    color: 'black',
  },
  keyLink: {
    display: 'flex',
    alignItems: 'center',
    width: 'fit-content',
  },
  spacing: {
    marginBottom: 96,
  },
})

const sortItems = (a: IntegrationKey, b: IntegrationKey): number => {
  const extA = a.externalSystemName || ''
  const extB = b.externalSystemName || ''
  if (extA.toLowerCase() < extB.toLowerCase()) return -1
  if (extA.toLowerCase() > extB.toLowerCase()) return 1
  if (extA < extB) return -1
  if (extA > extB) return 1
  if (a.name.toLowerCase() < b.name.toLowerCase()) return -1
  if (a.name.toLowerCase() > b.name.toLowerCase()) return 1
  if (a.name < b.name) return -1
  if (a.name > b.name) return 1
  return 0
}

export function IntegrationKeyDetails(props: {
  href: string
  label: string
  type: string
}): JSX.Element {
  const types = useFeatures().integrationKeyTypes
  const t = types.find((t) => t.id === props.type) || {
    enabled: false,
    name: props.type,
  }

  if (!t.enabled) {
    return (
      <React.Fragment>
        {t.name} integration keys are currently disabled.
      </React.Fragment>
    )
  }

  return <CopyText title={'Copy ' + props.label} value={props.href} asURL />
}

export default function IntegrationKeyList(props: {
  serviceID: string
}): JSX.Element {
  const classes = useStyles()
  const isMobile = useIsWidthDown('md')
  const [create, setCreate] = useState<boolean>(false)
  const [deleteDialog, setDeleteDialog] = useState<string | null>(null)

  const [{ fetching, error, data }] = useQuery({
    query,
    variables: { serviceID: props.serviceID },
  })
  const types = useFeatures().integrationKeyTypes
  const typeLabel = (type: string): string =>
    types.find((t) => t.id === type)?.label || type

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const items = (data.service.integrationKeys || [])
    .slice()
    .sort(sortItems)
    .filter((key: IntegrationKey) => !key.externalSystemName)
    .map(
      (key: IntegrationKey): FlatListListItem => ({
        title: key.name,
        subText: (
          <IntegrationKeyDetails
            key={key.id}
            href={key.href}
            label={typeLabel(key.type)}
            type={key.type}
          />
        ),
        secondaryAction: (
          <Grid container spacing={2} alignItems='center' wrap='nowrap'>
            {key.type === 'universal' && (
              <Grid item>
                <AppLink to={key.id}>
                  <Button
                    aria-label='Manage configuration and tokens for this key.'
                    onClick={() => {}}
                    variant='contained'
                  >
                    Manage
                  </Button>
                </AppLink>
              </Grid>
            )}
            <Grid item>
              <IconButton
                onClick={(): void => setDeleteDialog(key.id)}
                size='large'
              >
                <Trash />
              </IconButton>
            </Grid>
          </Grid>
        ),
      }),
    )

  const extItems = (data.service.integrationKeys || [])
    .slice()
    .sort(sortItems)
    .filter((key: IntegrationKey) => !!key.externalSystemName)
    .map(
      (key: IntegrationKey): FlatListListItem => ({
        title: key.name,
        subText: <Chip label={key.externalSystemName} sx={{ mt: 1 }} />,
        secondaryAction: (
          <IconButton
            onClick={(): void => setDeleteDialog(key.id)}
            size='large'
            sx={{ right: '-16px' }}
          >
            <Trash />
          </IconButton>
        ),
      }),
    )

  return (
    <React.Fragment>
      <Grid item xs={12} className={classes.spacing}>
        <Card>
          <CardContent>
            <FlatList
              data-cy='int-keys'
              headerNote={
                <React.Fragment>
                  API Documentation is available{' '}
                  <AppLink to='/docs'>here</AppLink>.
                </React.Fragment>
              }
              emptyMessage='No integration keys exist for this service.'
              items={items}
              headerAction={
                isMobile ? undefined : (
                  <Button
                    variant='contained'
                    onClick={(): void => setCreate(true)}
                    startIcon={<Add />}
                    data-testid='create-key'
                  >
                    Create Integration Key
                  </Button>
                )
              }
            />
            {!!extItems.length && (
              <Accordion disableGutters elevation={0}>
                <AccordionSummary
                  expandIcon={<ChevronDown />}
                  sx={{ ml: '6px', mr: '12px' }}
                >
                  <Typography>Externally Managed Keys</Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <FlatList
                    headerNote='These keys are managed by other applications.'
                    items={extItems}
                  />
                </AccordionDetails>
              </Accordion>
            )}
          </CardContent>
        </Card>
      </Grid>
      {isMobile && (
        <CreateFAB
          onClick={(): void => setCreate(true)}
          title='Create Integration Key'
        />
      )}
      <Suspense>
        {create && (
          <IntegrationKeyCreateDialog
            serviceID={props.serviceID}
            onClose={(): void => setCreate(false)}
          />
        )}
        {deleteDialog && (
          <IntegrationKeyDeleteDialog
            integrationKeyID={deleteDialog}
            onClose={(): void => setDeleteDialog(null)}
          />
        )}
      </Suspense>
    </React.Fragment>
  )
}
