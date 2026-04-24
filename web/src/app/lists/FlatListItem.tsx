import React, { useLayoutEffect } from 'react'
import ListItem from '@mui/material/ListItem'
import ListItemButton from '@mui/material/ListItemButton'
import ListItemIcon from '@mui/material/ListItemIcon'
import ListItemText from '@mui/material/ListItemText'
import AppLink, { AppLinkListItem } from '../util/AppLink'
import { FlatListItemOptions } from './FlatList'

const classes = {
  listItem: {
    width: '100%',
    marginTop: '8px',
    marginBottom: '8px',
    borderRadius: '4px',
  },
  listItemDisabled: {
    opacity: 0.6,
    width: '100%',
  },
  listItemDraggable: {
    paddingLeft: '75px',
  },
  secondaryText: {
    whiteSpace: 'pre-line',
  },
}

export interface FlatListItemProps {
  item: FlatListItemOptions
  index: number
}

export default function FlatListItem(
  props: FlatListItemProps,
): React.JSX.Element {
  const {
    highlight,
    selected,
    icon,
    secondaryAction,
    scrollIntoView,
    subText,
    title,
    url,
    draggable,
    disabled,
    disableTypography,
    onClick,
    primaryText,
    section,
    ...muiListItemProps
  } = props.item

  const ref = React.useRef<HTMLLIElement>(null)
  useLayoutEffect(() => {
    if (scrollIntoView) {
      ref.current?.scrollIntoView({ block: 'center' })
    }
  }, [scrollIntoView])

  const onClickProps = onClick && {
    onClick,
  }

  // When a URL is provided without a secondaryAction, AppLinkListItem renders
  // an <li> wrapper. In all other cases we need an explicit <ListItem> wrapper
  // so that children of a <ul>/<List> are valid <li> elements.
  const needsLiWrapper = !url || !!secondaryAction

  let linkProps = {}
  if (url) {
    linkProps = {
      component: secondaryAction ? AppLink : AppLinkListItem,
      to: url,
    }
  }

  const button = (
    <ListItemButton
      {...linkProps}
      {...onClickProps}
      {...muiListItemProps}
      sx={[
        !needsLiWrapper && classes.listItem,
        !!disabled && classes.listItemDisabled,
        !!draggable && classes.listItemDraggable,
      ]}
      selected={highlight}
    >
      {icon && <ListItemIcon tabIndex={-1}>{icon}</ListItemIcon>}
      <ListItemText
        primary={title || primaryText}
        secondary={subText}
        disableTypography={disableTypography}
        secondaryTypographyProps={{
          sx: [classes.secondaryText, !!disabled && classes.listItemDisabled],
          tabIndex: 0,
          component: typeof subText === 'string' ? 'p' : 'div',
        }}
      />
    </ListItemButton>
  )

  if (!needsLiWrapper) return button

  return (
    <ListItem
      ref={ref}
      disablePadding
      secondaryAction={secondaryAction}
      sx={classes.listItem}
    >
      {button}
    </ListItem>
  )
}
