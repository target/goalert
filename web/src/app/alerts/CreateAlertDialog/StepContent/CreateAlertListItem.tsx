import React from 'react'
import { ListItem, ListItemText, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import CopyText from '../../../util/CopyText'
import AppLink from '../../../util/AppLink'

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

export default function CreateAlertListItem(props: {
  id: string
}): React.JSX.Element {
  const { id } = props

  const classes = useStyles()

  const alertURL = '/alerts/' + id

  return (
    <ListItem key={id} divider>
      <ListItemText disableTypography className={classes.listItemText}>
        <span>
          <Typography>
            <AppLink to={alertURL} newTab>
              #{id}
            </AppLink>
          </Typography>
        </span>

        <span className={classes.endLinks}>
          <CopyText value={alertURL} placement='left' asURL />
          <AppLink to={alertURL} newTab className={classes.openInNewTab}>
            <OpenInNewIcon fontSize='small' />
          </AppLink>
        </span>
      </ListItemText>
    </ListItem>
  )
}
