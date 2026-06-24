import { List, ListItem, ListItemText, Typography } from '@mui/material'
import React from 'react'
import ReorderGroup from './ReorderGroup'
import { has } from 'lodash'
import { useIsWidthDown } from '../util/useWidth'

export type CompListProps = {
  note?: React.ReactNode
  action?: React.ReactNode

  /* If true, the list will only show the action on wide screens. */
  hideActionOnMobile?: boolean
  emptyMessage?: string
  children?: React.ReactNode

  /* data-cy attribute for testing */
  'data-cy'?: string
}

function isReorderGroup(child: React.ReactNode): child is React.ReactElement {
  return (
    React.isValidElement(child) &&
    has(child.props, 'children') &&
    child.type === ReorderGroup
  )
}

/* A composable list component. */
export default function CompList(props: CompListProps): React.ReactNode {
  const children = React.Children.toArray(props.children)
  let hasNoChildren = children.length === 0
  const isMobile = useIsWidthDown('md')

  // Special case: ReorderGroup with no contents/children as a child
  if (
    children.length > 0 &&
    children.every(
      (child) =>
        isReorderGroup(child) &&
        React.Children.count(child.props.children) === 0,
    )
  ) {
    hasNoChildren = true
  }

  const emptyMessage = props.emptyMessage ?? 'No results.'
  return (
    <List data-cy={props['data-cy']} sx={{ display: 'grid' }}>
      {(props.note || props.action) && (
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
          {props.action && (props.hideActionOnMobile ? !isMobile : true) && (
            <div>{props.action}</div>
          )}
        </ListItem>
      )}
      {!hasNoChildren
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
