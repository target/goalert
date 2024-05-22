import React, { MouseEventHandler } from 'react'
import makeStyles from '@mui/styles/makeStyles'
import Button, { ButtonProps } from '@mui/material/Button'
import MUICardActions from '@mui/material/CardActions'
import IconButton from '@mui/material/IconButton'
import Tooltip from '@mui/material/Tooltip'
import { Theme } from '@mui/material'

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
  ButtonProps?: ButtonProps
}

const useStyles = makeStyles((theme: Theme) => ({
  cardActions: {
    alignItems: 'flex-end', // aligns icon buttons to bottom of container
  },
  primaryActionsContainer: {
    padding: 8,
    width: '100%',
    '& > button:not(:last-child)': {
      marginRight: theme.spacing(2),
    },
  },
  autoExpandWidth: {
    margin: '0 auto',
  },
}))

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
    // wrapping button in span so tooltip can still
    // render when hovering over disabled buttons
    return (
      <Tooltip title={action.label} placement='top'>
        <span aria-label={undefined}>
          <IconButton
            aria-label={action.label}
            onClick={action.handleOnClick}
            size='large'
            {...action.ButtonProps}
          >
            {action.icon}
          </IconButton>
        </span>
      </Tooltip>
    )
  }

  return (
    <Button
      onClick={action.handleOnClick}
      startIcon={action.icon}
      variant='contained'
      {...action.ButtonProps}
    >
      {action.label}
    </Button>
  )
}
