import React from 'react'
import { PropTypes as p } from 'prop-types'
import Tooltip from '@material-ui/core/Tooltip'
import AddIcon from '@material-ui/icons/Add'
import TrashIcon from '@material-ui/icons/Delete'
import WarningIcon from '@material-ui/icons/Warning'
import slackIcon from '../../public/slack.svg'
import slackIconBW from '../../public/slack_monochrome_black.svg'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles({
  trashIcon: {
    color: '#666',
    cursor: 'pointer',
    float: 'right',
  },
  warningColor: {
    color: '#FFD602',
  },
})

export function Trash() {
  const classes = useStyles()

  return <TrashIcon className={classes.trashIcon} />
}

export function Warning(props) {
  const { message } = props
  const classes = useStyles()

  const warningIcon = (
    <WarningIcon data-cy='warning-icon' className={classes.warningColor} />
  )

  if (!message) {
    return warningIcon
  }

  return (
    <Tooltip title={message} placement='right'>
      {warningIcon}
    </Tooltip>
  )
}

export function Add() {
  return <AddIcon />
}

export function Slack() {
  return <img src={slackIcon} width={20} height={20} alt='Slack' />
}

export function SlackBW() {
  return <img src={slackIconBW} width={20} height={20} alt='Slack' />
}

Warning.propTypes = {
  message: p.string,
}
