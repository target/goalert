import React from 'react'
import p from 'prop-types'
import {
  ListItem,
  ListItemText,
  Typography,
  Link,
  makeStyles,
} from '@material-ui/core'
import OpenInNewIcon from '@material-ui/icons/OpenInNew'
import CopyText from '../../../util/CopyText'
import { absURLSelector } from '../../../selectors'
import { useSelector } from 'react-redux'

const useStyles = makeStyles({
  listItemText: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  endLinks: {
    display: 'flex',
    alignItems: 'flex-start',
  },
  openInNewTab: {
    marginLeft: '0.75em',
  },
})

export default function CreateAlertListItem(props) {
  const { id } = props

  const classes = useStyles()

  const absURL = useSelector(absURLSelector)
  const alertURL = window.location.origin + absURL('/alerts/' + id)

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

        <span className={classes.endLinks}>
          <CopyText value={alertURL} placement='left' />
          <a
            href={alertURL}
            target='_blank'
            rel='noopener noreferrer'
            className={classes.openInNewTab}
          >
            <OpenInNewIcon fontSize='small' />
          </a>
        </span>
      </ListItemText>
    </ListItem>
  )
}

CreateAlertListItem.propTypes = {
  id: p.string.isRequired,
}
