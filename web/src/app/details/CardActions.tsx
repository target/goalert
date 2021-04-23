import React, { MouseEventHandler } from 'react'
import { makeStyles } from '@material-ui/core/styles'
import Button from '@material-ui/core/Button'
import MUICardActions from '@material-ui/core/CardActions'
import IconButton from '@material-ui/core/IconButton'
import Tooltip from '@material-ui/core/Tooltip'

interface CardActionProps {
  primaryActions?: Array<Action | JSX.Element>
  secondaryActions?: Array<Action | JSX.Element>
}

export type Action = {
  label: string // primary button text, use for a tooltip if secondary action
  handleOnClick: MouseEventHandler<HTMLButtonElement>

  icon?: JSX.Element // if true, adds a start icon to a button with text
  secondary?: boolean // if true, renders right-aligned as an icon button
}

const useStyles = makeStyles({
  cardActions: {
    alignItems: 'flex-end', // aligns icon buttons to bottom of container
  },
})

export default function CardActions(p: CardActionProps): JSX.Element {
  const classes = useStyles()

  const action = (action: Action | JSX.Element, key: string): JSX.Element => {
    if ('label' in action && 'handleOnClick' in action) {
      return <Action key={key} {...action} />
    }
    return action
  }

  let actions: Array<JSX.Element> = []
  if (p.primaryActions) {
    actions = p.primaryActions.map((a, i) => action(a, 'primary' + i))
  }
  if (p.secondaryActions) {
    actions = [
      ...actions,
      <div key='actions-margin' style={{ margin: '0 auto' }} />,
      ...p.secondaryActions.map((a, i) =>
        action({ ...a, secondary: true }, 'secondary' + i),
      ),
    ]
  }

  return (
    <MUICardActions data-cy='card-actions' className={classes.cardActions}>
      {actions}
    </MUICardActions>
  )
}

function Action(p: Action): JSX.Element {
  if (p.secondary && p.icon) {
    return (
      <Tooltip title={p.label} placement='top'>
        <IconButton onClick={p.handleOnClick}>{p.icon}</IconButton>
      </Tooltip>
    )
  }
  return (
    <Button onClick={p.handleOnClick} startIcon={p.icon}>
      {p.label}
    </Button>
  )
}
