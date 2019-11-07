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
import ContentCopyIcon from 'mdi-material-ui/ContentCopy'

import gql from 'graphql-tag'
import copyToClipboard from '../../util/copyToClipboard-v2'

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

  const alertUrl = `${window.location.origin}/alerts/${alert.id}`

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
          {/* <Typography>
            <b>{alert.id}: </b>
            {alert.status.toUpperCase().slice(6)}
          </Typography>
          <Typography variant='caption'>{alert.service.name}</Typography> */}
          <Typography>
            <Link href={alertUrl} target='_blank' rel='noopener noreferrer'>
              {alertUrl}
            </Link>
          </Typography>
        </span>

        <span>
          <IconButton
            aria-label='Copy alert URL'
            onClick={e => {
              copyToClipboard(alertUrl)
            }}
          >
            <ContentCopyIcon fontSize='small' />
          </IconButton>
          <IconButton
            aria-label='Open alert in new tab'
            onClick={() => window.open(alertUrl)}
          >
            <OpenInNewIcon fontSize='small' />
          </IconButton>
        </span>
      </ListItemText>
    </ListItem>
  )
}

AlertListItem.propTypes = {
  id: p.string.isRequired,
}
