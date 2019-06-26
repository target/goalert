import React, { Component } from 'react'
import Tooltip from '@material-ui/core/Tooltip'
import withStyles from '@material-ui/core/styles/withStyles'
import AddIcon from '@material-ui/icons/Add'
import TrashIcon from '@material-ui/icons/Delete'
import WarningIcon from '@material-ui/icons/Warning'
import except from 'except'
import { styles } from '../../styles/materialStyles'
import slackIcon from '../../public/slack.svg'
import slackIconBW from '../../public/slack_monochrome_black.svg'

@withStyles(styles)
export class Trash extends Component {
  render() {
    const { classes } = this.props

    return (
      <TrashIcon
        className={classes.trashIcon}
        {...except(this.props, 'classes')}
      />
    )
  }
}

@withStyles(styles)
export class Warning extends Component {
  render() {
    const { classes, details } = this.props

    const warningIcon = (
      <WarningIcon
        className={classes.warningColor}
        {...except(this.props, 'classes')}
      />
    )

    if (!details) {
      return warningIcon
    }

    return (
      <Tooltip title={details} placement='right'>
        {warningIcon}
      </Tooltip>
    )
  }
}

@withStyles(styles)
export class Add extends Component {
  render() {
    return <AddIcon {...except(this.props, 'classes')} />
  }
}

@withStyles(styles)
export class Slack extends Component {
  render() {
    return (
      <img
        src={slackIcon}
        {...except(this.props, 'classes')}
        width={20}
        height={20}
        alt='Slack'
      />
    )
  }
}

export class SlackBW extends Component {
  render() {
    return (
      <img
        src={slackIconBW}
        {...except(this.props, 'classes')}
        width={20}
        height={20}
        alt='Slack'
      />
    )
  }
}
