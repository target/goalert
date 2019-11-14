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

import copyToClipboard from '../../../util/copyToClipboard'
import { absURLSelector } from '../../../selectors'
import { useSelector } from 'react-redux'

const useStyles = makeStyles(theme => ({
  listItemText: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
}))

export default function CreateAlertListItem(props) {
  const { id } = props

  const classes = useStyles()

  const absURL = useSelector(absURLSelector)
  const alertURL = absURL('/alerts/' + id)

  return (
    <ListItem key={id} divider>
      <ListItemText disableTypography className={classes.listItemText}>
        <span>
          <Typography>
            <Link href={alertURL} target='_blank' rel='noopener noreferrer'>
              #{id}
            </Link>
          </Typography>
        </span>

        <span>
          <IconButton
            aria-label='Copy alert URL'
            onClick={e => {
              copyToClipboard(alertURL)
            }}
          >
            <ContentCopyIcon fontSize='small' />
          </IconButton>
          <Link href={alertURL} target='_blank' rel='noopener noreferrer'>
            <IconButton aria-label='Open alert in new tab'>
              <OpenInNewIcon fontSize='small' />
            </IconButton>
          </Link>
        </span>
      </ListItemText>
    </ListItem>
  )
}

CreateAlertListItem.propTypes = {
  id: p.string.isRequired,
}
