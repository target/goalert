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

interface ActionProps {
  action: Action
  secondary?: boolean // if true, renders right-aligned as an icon button
}

export type Action = {
  label: string // primary button text, use for a tooltip if secondary action
  handleOnClick: MouseEventHandler<HTMLButtonElement>

  icon?: JSX.Element // if true, adds a start icon to a button with text
}

const useStyles = makeStyles({
  cardActions: {
    alignItems: 'flex-end', // aligns icon buttons to bottom of container
  },
  primaryActionsContainer: {
    padding: 8,
  },
  autoExpandWidth: {
    margin: '0 auto',
  },
})

export default function CardActions(p: CardActionProps): JSX.Element {
  const classes = useStyles()

  const action = (
    action: Action | JSX.Element,
    key: string,
    secondary?: boolean,
  ): JSX.Element => {
    if ('label' in action && 'handleOnClick' in action) {
      return <Action key={key} action={action} secondary={secondary} />
    }
    return action
  }

  let actions: Array<JSX.Element> = []
  if (p.primaryActions) {
    actions = [
      <div
        key='primary-actions-container'
        className={classes.primaryActionsContainer}
      >
        {p.primaryActions.map((a, i) => action(a, 'primary' + i))}
      </div>,
    ]
  }
  if (p.secondaryActions) {
    actions = [
      ...actions,
      <div key='actions-margin' className={classes.autoExpandWidth} />,
      ...p.secondaryActions.map((a, i) => action(a, 'secondary' + i, true)),
    ]
  }

  return (
    <MUICardActions data-cy='card-actions' className={classes.cardActions}>
      {actions}
    </MUICardActions>
  )
}

function Action(p: ActionProps): JSX.Element {
  const { action, secondary } = p
  if (secondary && action.icon) {
    return (
      <Tooltip title={action.label} placement='top'>
        <IconButton onClick={action.handleOnClick}>{action.icon}</IconButton>
      </Tooltip>
    )
  }
  return (
    <Button onClick={action.handleOnClick} startIcon={action.icon}>
      {action.label}
    </Button>
  )
}
