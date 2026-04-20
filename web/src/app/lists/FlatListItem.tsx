import React, { useLayoutEffect } from 'react'
import classnames from 'classnames'
import ListItemButton from '@mui/material/ListItemButton'
import ListItemIcon from '@mui/material/ListItemIcon'
import ListItemSecondaryAction from '@mui/material/ListItemSecondaryAction'
import ListItemText from '@mui/material/ListItemText'
import makeStyles from '@mui/styles/makeStyles'
import AppLink, { AppLinkListItem } from '../util/AppLink'
import { FlatListItemOptions } from './FlatList'

const useStyles = makeStyles(() => ({
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
}))

export interface FlatListItemProps {
  item: FlatListItemOptions
  index: number
}

export default function FlatListItem(
  props: FlatListItemProps,
): React.JSX.Element {
  const classes = useStyles()

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

  let linkProps = {}
  if (url) {
    linkProps = {
      // if you render a link with a secondary action, MUI will render the <a> tag without an <li> around it
      component: secondaryAction ? AppLink : AppLinkListItem,
      to: url,
    }
  }

  const onClickProps = onClick && {
    onClick,
  }

  return (
    <ListItemButton
      {...linkProps}
      {...onClickProps}
      {...muiListItemProps}
      className={classnames({
        [classes.listItem]: true,
        [classes.listItemDisabled]: disabled,
        [classes.listItemDraggable]: draggable,
      })}
      selected={highlight}
    >
      {icon && <ListItemIcon tabIndex={-1}>{icon}</ListItemIcon>}
      <ListItemText
        primary={title || primaryText}
        secondary={subText}
        disableTypography={disableTypography}
        secondaryTypographyProps={{
          className: classnames({
            [classes.secondaryText]: true,
            [classes.listItemDisabled]: disabled,
          }),
          tabIndex: 0,
          component: typeof subText === 'string' ? 'p' : 'div',
        }}
      />
      {secondaryAction && (
        <ListItemSecondaryAction>{secondaryAction}</ListItemSecondaryAction>
      )}
    </ListItemButton>
  )
}
