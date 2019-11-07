import React from 'react'
import { PropTypes as p } from 'prop-types'
import { useQuery } from 'react-apollo'
import {
  ListItem,
  ListItemText,
  Typography,
  IconButton,
  Link,
} from '@material-ui/core'
import OpenInNewIcon from '@material-ui/icons/OpenInNew'

import gql from 'graphql-tag'

const serviceQuery = gql`
  query service($id: ID!) {
    service(id: $id) {
      id
      name
    }
  }
`

export default function ServiceListItem(props) {
  const { id, err } = props

  const { data, loading, error } = useQuery(serviceQuery, {
    variables: {
      id,
    },
  })

  let { service } = data || {}

  if (loading) return 'Loading...'
  if (error) return 'Error fetching data.'

  const serviceUrl = `${window.location.origin}/services/${id}`

  return (
    <ListItem key={id} divider>
      <ListItemText
        disableTypography
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <span>
          <Typography>
            <Link href={serviceUrl} target='_blank' rel='noopener noreferrer'>
              {service.name}
            </Link>
          </Typography>
          <Typography color='error' variant='caption'>
            {err}
          </Typography>
        </span>

        <Link href={serviceUrl} target='_blank' rel='noopener noreferrer'>
          <IconButton aria-label='Open service in new tab'>
            <OpenInNewIcon fontSize='small' />
          </IconButton>
        </Link>
      </ListItemText>
    </ListItem>
  )
}

ServiceListItem.propTypes = {
  id: p.string.isRequired,
}
