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

export default function FlatListItem(props: FlatListItemProps): JSX.Element {
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

  return (
    <MUIListItem
      key={props.index}
      {...linkProps}
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
        primary={title}
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
