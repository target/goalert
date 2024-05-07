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

export interface FlatListItemProps extends MUIListItemProps {
  item: FlatListItemOptions
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
      // if you render a link with a secondary action, MUI will render the <a> tag without an <li> around it
      component: secondaryAction ? AppLink : 'li',
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
          component: typeof subText === 'string' ? 'p' : 'div',
        }}
      />
      {secondaryAction && (
        <ListItemSecondaryAction>{secondaryAction}</ListItemSecondaryAction>
      )}
    </MUIListItem>
  )
}
