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
  secondaryText: {
    whiteSpace: 'pre-line',
  },
}))

interface ScrollIntoViewListItemProps extends MUIListItemProps {
  scrollIntoView?: boolean
}

function ScrollIntoViewListItem(
  props: ScrollIntoViewListItemProps,
): JSX.Element {
  const { scrollIntoView, ...other } = props
  const ref = React.useRef<HTMLLIElement>(null)
  useLayoutEffect(() => {
    if (scrollIntoView) {
      ref.current?.scrollIntoView({ block: 'center' })
    }
  }, [scrollIntoView])

  return <MUIListItem ref={ref} {...other} />
}

interface ListItemProps extends MUIListItemProps {
  item: FlatListItem
  index: number
}

export default function ListItem({ item, index }: ListItemProps): JSX.Element {
  const classes = useStyles()

  let linkProps = {}
  if (item.url) {
    linkProps = {
      component: AppLink,
      to: item.url,
      button: true,
    }
  }

  return (
    <ScrollIntoViewListItem
      scrollIntoView={item.scrollIntoView}
      key={index}
      {...linkProps}
      className={classnames({
        [classes.listItem]: true,
        [classes.listItemDisabled]: item.disabled,
      })}
      selected={item.highlight}
    >
      {item.icon && <ListItemIcon>{item.icon}</ListItemIcon>}
      <ListItemText
        primary={item.title}
        secondary={item.subText}
        secondaryTypographyProps={{
          className: classnames({
            [classes.secondaryText]: true,
            [classes.listItemDisabled]: item.disabled,
          }),
        }}
      />
      {item.secondaryAction && (
        <ListItemSecondaryAction>
          {item.secondaryAction}
        </ListItemSecondaryAction>
      )}
    </ScrollIntoViewListItem>
  )
}
