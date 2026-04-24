import React from 'react'
import { ListItem, ListItemText, Typography } from '@mui/material'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import CopyText from '../../../util/CopyText'
import AppLink from '../../../util/AppLink'

export default function CreateAlertListItem(props: {
  id: string
}): React.JSX.Element {
  const { id } = props

  const alertURL = '/alerts/' + id

  return (
    <ListItem key={id} divider>
      <ListItemText
        disableTypography
        sx={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <span>
          <Typography>
            <AppLink to={alertURL} newTab>
              #{id}
            </AppLink>
          </Typography>
        </span>

        <span style={{ display: 'flex', alignItems: 'flex-start' }}>
          <CopyText value={alertURL} placement='left' asURL />
          <AppLink to={alertURL} newTab style={{ marginLeft: '0.75em' }}>
            <OpenInNewIcon fontSize='small' />
          </AppLink>
        </span>
      </ListItemText>
    </ListItem>
  )
}
