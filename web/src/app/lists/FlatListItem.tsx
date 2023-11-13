import React, { useLayoutEffect } from 'react'
import classnames from 'classnames'
import MUIListItem, {
  ListItemProps as MUIListItemProps,
} from '@mui/material/ListItem'
import ListItemIcon from '@mui/material/ListItemIcon'
import ListItemSecondaryAction from '@mui/material/ListItemSecondaryAction'
import ListItemText from '@mui/material/ListItemText'
import makeStyles from '@mui/styles/makeStyles'
import AppLink from '../util/AppLink'
import { FlatListItem } from './FlatList'

const useStyles = makeStyles(() => ({
  listItem: {
    width: '100%',
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

export interface FlatListItemProps extends MUIListItemProps {
  item: FlatListItem
  index: number
}

export default function FlatListItem(props: FlatListItemProps): React.ReactNode {
  const classes = useStyles()

  const {
    highlight,
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
      component: AppLink,
      to: url,
      button: true,
    }
  }

  const onClickProps = onClick && {
    onClick,

    // NOTE: needed for error: button: false? not assignable to type 'true'
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    button: true as any,
  }

  return (
    <MUIListItem
      key={props.index}
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
        }}
      />
      {secondaryAction && (
        <ListItemSecondaryAction>{secondaryAction}</ListItemSecondaryAction>
      )}
    </MUIListItem>
  )
}
