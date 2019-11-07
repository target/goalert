import React from 'react'
import { PropTypes as p } from 'prop-types'
import {
  ListItem,
  ListItemText,
  Typography,
  IconButton,
  Link,
} from '@material-ui/core'
import OpenInNewIcon from '@material-ui/icons/OpenInNew'
import ContentCopyIcon from 'mdi-material-ui/ContentCopy'

import copyToClipboard from '../../util/copyToClipboard-v2'

export default function AlertListItem(props) {
  const { id } = props

  const alertUrl = `${window.location.origin}/alerts/${id}`

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
