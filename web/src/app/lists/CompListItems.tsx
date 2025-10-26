import {
  Collapse,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemSecondaryAction,
  ListItemText,
} from '@mui/material'
import React from 'react'
import AppLink from '../util/AppLink'
import { ExpandLess, ExpandMore } from '@mui/icons-material'

export type CompListItemTextProps = {
  title?: React.ReactNode
  icon?: React.ReactNode
  /* Set to true to always create space for the icon, even if it is not set */
  alwaysShowIcon?: boolean
  highlight?: boolean

  subText?: React.ReactNode
  action?: React.ReactNode
  disableTypography?: boolean
}

/* A simple list item with a title, subtext, and optional action */
export function CompListItemText(
  props: CompListItemTextProps,
): React.ReactNode {
  return (
    <ListItem className={props.highlight ? 'Mui-selected' : ''}>
      {(props.icon || props.alwaysShowIcon) && (
        <ListItemIcon tabIndex={-1}>{props.icon}</ListItemIcon>
      )}
      <ListItemText
        disableTypography={props.disableTypography}
        primary={props.title}
        secondary={props.subText}
      />
      {props.action && (
        <ListItemSecondaryAction>{props.action}</ListItemSecondaryAction>
      )}
    </ListItem>
  )
}

export type CompListItemNavProps = CompListItemTextProps & {
  url: string
}

/* A list item that links to a URL. */
export function CompListItemNav(props: CompListItemNavProps): React.ReactNode {
  return (
    <li>
      <ListItemButton component={AppLink} to={props.url}>
        {props.icon && <ListItemIcon tabIndex={-1}>{props.icon}</ListItemIcon>}
        <ListItemText primary={props.title} secondary={props.subText} />
        {props.action && (
          <ListItemSecondaryAction>{props.action}</ListItemSecondaryAction>
        )}
      </ListItemButton>
    </li>
  )
}

export type CompListSectionProps = {
  title: React.ReactNode
  subText?: React.ReactNode
  icon?: React.ReactNode
  children?: React.ReactNode
  defaultOpen?: boolean
}

/* A collapsible section of list items. */
export function CompListSection(props: CompListSectionProps): React.ReactNode {
  const [open, setOpen] = React.useState(props.defaultOpen ?? false)

  return (
    <React.Fragment>
      <ListItemButton onClick={() => setOpen(!open)}>
        {props.icon && <ListItemIcon>{props.icon}</ListItemIcon>}
        <ListItemText primary={props.title} secondary={props.subText} />
        {open ? <ExpandLess /> : <ExpandMore />}
      </ListItemButton>
      <Collapse in={open} unmountOnExit>
        <List>{props.children}</List>
      </Collapse>
    </React.Fragment>
  )
}
