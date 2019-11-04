import React from 'react'
import { PropTypes as p } from 'prop-types'
import { useQuery } from 'react-apollo'
import { ListItem, ListItemText, Typography } from '@material-ui/core'
import { Link } from 'react-router-dom'

import gql from 'graphql-tag'

const alertQuery = gql`
  query alert($id: Int!) {
    alert(id: $id) {
      id
      summary
      status
      service {
        name
      }
    }
  }
`

export default function AlertListItem(props) {
  const { id } = props

  const { data, loading, error } = useQuery(alertQuery, {
    variables: {
      id,
    },
  })

  let { alert } = data || {}

  if (loading) return 'Loading...'
  if (error) return 'Error fetching data.'

  return (
    <ListItem
      button
      key={id}
      component={Link}
      to={`/alerts/${alert.id}`}
      target={'_blank'}
    >
      <ListItemText disableTypography style={{ paddingRight: '2.75em' }}>
        <Typography>
          <b>{alert.id}: </b>
          {alert.status.toUpperCase().slice(6)}
        </Typography>
        <Typography variant='caption'>{alert.service.name}</Typography>
      </ListItemText>
    </ListItem>
  )
}

AlertListItem.propTypes = {
  id: p.string.isRequired,
}
