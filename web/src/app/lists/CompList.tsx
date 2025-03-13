import { List, ListItem, ListItemText, Typography } from '@mui/material'
import React from 'react'

export type CompListProps = {
  note?: React.ReactNode
  action?: React.ReactNode
  emptyMessage?: string
  children?: React.ReactNode
}

/* A composable list component. */
export default function CompList(props: CompListProps): React.ReactNode {
  const emptyMessage = props.emptyMessage ?? 'No results.'
  return (
    <List>
      <ListItem>
        {props.note && (
          <ListItemText
            disableTypography
            secondary={
              <Typography color='textSecondary'>{props.note}</Typography>
            }
            sx={{ fontStyle: 'italic', pr: 2 }}
          />
        )}
        {props.action && <div>{props.action}</div>}
      </ListItem>
      {React.Children.count(props.children)
        ? props.children
        : emptyMessage && (
            <ListItem>
              <ListItemText
                disableTypography
                secondary={
                  <Typography data-cy='list-empty-message' variant='caption'>
                    {emptyMessage}
                  </Typography>
                }
              />
            </ListItem>
          )}
    </List>
  )
}
