import React, { useState } from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import { makeStyles } from '@material-ui/core'
import { gql } from '@apollo/client'
import CreateFAB from '../lists/CreateFAB'
import FlatList from '../lists/FlatList'
import Query from '../util/Query'
import OtherActions from '../util/OtherActions'

import ServiceLabelSetDialog from './ServiceLabelCreateDialog'
import ServiceLabelEditDialog from './ServiceLabelEditDialog'
import ServiceLabelDeleteDialog from './ServiceLabelDeleteDialog'

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

const styles = () => ({
  spacing: { marginBottom: 96 },
})

const sortItems = (a, b) => {
  if (a.key.toLowerCase() < b.key.toLowerCase()) return -1
  if (a.key.toLowerCase() > b.key.toLowerCase()) return 1
  if (a.key < b.key) return -1
  if (a.key > b.key) return 1
  return 0
}

const useStyles = makeStyles(styles)

export default function ServiceLabelList(props) {
  const { serviceID } = props
  const [create, setCreate] = useState(false)
  const [editKey, setEditKey] = useState(null)
  const [deleteKey, setDeleteKey] = useState(null)
  const classes = useStyles()

  function renderList(labels) {
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

  function renderQuery() {
    return (
      <Query
        query={query}
        variables={{ serviceID: serviceID }}
        render={({ data }) => renderList(data.service.labels)}
      />
    )
  }

  return (
    <React.Fragment>
      <Grid item xs={12} className={classes.spacing}>
        <Card>
          <CardContent>{renderQuery()}</CardContent>
        </Card>
      </Grid>
      <CreateFAB onClick={() => setCreate(true)} title='Add Label' />
      {create && (
        <ServiceLabelSetDialog
          serviceID={serviceID}
          onClose={() => setCreate(false)}
        />
      )}
      {editKey && (
        <ServiceLabelEditDialog
          serviceID={serviceID}
          labelKey={editKey}
          onClose={() => setEditKey(null)}
        />
      )}
      {deleteKey && (
        <ServiceLabelDeleteDialog
          serviceID={serviceID}
          labelKey={deleteKey}
          onClose={() => setDeleteKey(null)}
        />
      )}
    </React.Fragment>
  )
}

ServiceLabelList.propTypes = {
  serviceID: p.string.isRequired,
}
