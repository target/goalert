import React, { useState } from 'react'
import Grid from '@mui/material/Grid'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import makeStyles from '@mui/styles/makeStyles'
import { gql, useQuery } from 'urql'
import CreateFAB from '../lists/CreateFAB'
import FlatList from '../lists/FlatList'
import OtherActions from '../util/OtherActions'

import ServiceLabelSetDialog from './ServiceLabelCreateDialog'
import ServiceLabelEditDialog from './ServiceLabelEditDialog'
import ServiceLabelDeleteDialog from './ServiceLabelDeleteDialog'
import { Label } from '../../schema'
import Spinner from '../loading/components/Spinner'

const query = gql`
  query ($serviceID: ID!) {
    service(id: $serviceID) {
      id # need to tie the result to the correct record
      labels {
        key
        value
      }
    }
  }
`

const sortItems = (a: Label, b: Label): number => {
  if (a.key.toLowerCase() < b.key.toLowerCase()) return -1
  if (a.key.toLowerCase() > b.key.toLowerCase()) return 1
  if (a.key < b.key) return -1
  if (a.key > b.key) return 1
  return 0
}

const useStyles = makeStyles({ spacing: { marginBottom: 96 } })

export default function ServiceLabelList(props: { serviceID: string }) {
  const [create, setCreate] = useState(false)
  const [editKey, setEditKey] = useState<string | null>(null)
  const [deleteKey, setDeleteKey] = useState<string | null>(null)
  const classes = useStyles()

  const [{ data, fetching }] = useQuery({
    query,
    variables: { serviceID: props.serviceID },
  })

  if (!data && fetching) {
    return <Spinner />
  }

  function renderList(labels: Label[]) {
    const items = (labels || [])
      .slice()
      .sort(sortItems)
      .map((label) => ({
        title: label.key,
        subText: label.value,
        secondaryAction: (
          <OtherActions
            actions={[
              {
                label: 'Edit',
                onClick: () => setEditKey(label.key),
              },
              {
                label: 'Delete',
                onClick: () => setDeleteKey(label.key),
              },
            ]}
          />
        ),
      }))

    return (
      <FlatList
        data-cy='label-list'
        emptyMessage='No labels exist for this service.'
        items={items}
      />
    )
  }

  return (
    <React.Fragment>
      <Grid item xs={12} className={classes.spacing}>
        <Card>
          <CardContent>{renderList(data.service.labels)}</CardContent>
        </Card>
      </Grid>
      <CreateFAB onClick={() => setCreate(true)} title='Add Label' />
      {create && (
        <ServiceLabelSetDialog
          serviceID={props.serviceID}
          onClose={() => setCreate(false)}
        />
      )}
      {editKey && (
        <ServiceLabelEditDialog
          serviceID={props.serviceID}
          labelKey={editKey}
          onClose={() => setEditKey(null)}
        />
      )}
      {deleteKey && (
        <ServiceLabelDeleteDialog
          serviceID={props.serviceID}
          labelKey={deleteKey}
          onClose={() => setDeleteKey(null)}
        />
      )}
    </React.Fragment>
  )
}
