import React from 'react'
import { PropTypes as p } from 'prop-types'
import {
  ListItem,
  ListItemText,
  Typography,
  IconButton,
  Link,
  makeStyles,
} from '@material-ui/core'
import OpenInNewIcon from '@material-ui/icons/OpenInNew'
import ContentCopyIcon from 'mdi-material-ui/ContentCopy'

import copyToClipboard from '../../util/copyToClipboard-v2'
import { absURLSelector } from '../../selectors'

const useStyles = makeStyles(theme => ({
  listItemText: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
}))

export default function AlertListItem(props) {
  const { id } = props

  const classes = useStyles()

  const selectAlertUrl = absURLSelector({
    router: { location: { pathname: 'alerts' } },
  })

  const alertUrl = `${window.location.origin}/${selectAlertUrl(id)}`

  return (
    <ListItem key={id} divider>
      <ListItemText disableTypography className={classes.listItemText}>
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
          <Link href={alertUrl} target='_blank' rel='noopener noreferrer'>
            <IconButton aria-label='Open alert in new tab'>
              <OpenInNewIcon fontSize='small' />
            </IconButton>
          </Link>
        </span>
      </ListItemText>
    </ListItem>
  )
}

AlertListItem.propTypes = {
  id: p.string.isRequired,
}
