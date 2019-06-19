import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Chip from '@material-ui/core/Chip'
import { withRouter } from 'react-router-dom'

import {
  Layers as PolicyIcon,
  RotateRight as RotationIcon,
  Today as ScheduleIcon,
  VpnKey as ServiceIcon,
} from '@material-ui/icons'
import Avatar from '@material-ui/core/Avatar'
import { UserAvatar } from './avatar'
import { SlackBW } from '../icons'

@withRouter
export class ServiceChip extends Component {
  static propTypes = {
    id: p.string.isRequired,
    style: p.object,
    name: p.string.isRequired,
    onDelete: p.func,
  }

  render() {
    const { id, history, name, onDelete, style } = this.props

    return (
      <Chip
        data-cy='service-chip'
        avatar={
          <Avatar>
            <ServiceIcon />
          </Avatar>
        }
        style={style}
        onDelete={onDelete}
        onClick={() => history.push(`/services/${id}`)}
        label={name}
      />
    )
  }
}

@withRouter
export class UserChip extends Component {
  static propTypes = {
    id: p.string.isRequired,
    name: p.string.isRequired,
    onDelete: p.func,
    onClick: p.func,
    style: p.object,
  }

  render() {
    const { id, history, name, onDelete, onClick, style } = this.props

    let localOnClick = () => history.push(`/users/${id}`)
    if (onClick) {
      localOnClick = onClick
    }

    return (
      <Chip
        data-cy='user-chip'
        avatar={<UserAvatar userID={id} />}
        onDelete={onDelete}
        onClick={localOnClick}
        label={name}
        style={style}
      />
    )
  }
}

@withRouter
export class RotationChip extends Component {
  static propTypes = {
    id: p.string.isRequired,
    style: p.object,
    name: p.string.isRequired,
    onDelete: p.func,
  }

  render() {
    const { id, history, name, onDelete, style } = this.props

    return (
      <Chip
        data-cy='rotation-chip'
        avatar={
          <Avatar>
            <RotationIcon />
          </Avatar>
        }
        onDelete={onDelete}
        onClick={() => history.push(`/rotations/${id}`)}
        label={name}
        style={style}
      />
    )
  }
}

const formatPolicyName = (name, stepNum) => {
  if (stepNum !== null) {
    return (
      <div>
        <strong> Step {stepNum + 1}:</strong> {name}
      </div>
    )
  }

  const parts = name.split(' - ')
  return (
    <div>
      <strong> Step {parseInt(parts[0]) + 1}:</strong> {parts[1]}
    </div>
  )
}

@withRouter
export class PolicyChip extends Component {
  static propTypes = {
    id: p.string.isRequired,
    style: p.object,
    name: p.string.isRequired,
    stepNum: p.number,
    onDelete: p.func,
  }

  render() {
    const { id, history, name, stepNum, onDelete, style } = this.props

    return (
      <Chip
        data-cy='ep-chip'
        avatar={
          <Avatar>
            <PolicyIcon />
          </Avatar>
        }
        onDelete={onDelete}
        onClick={() => history.push(`/escalation-policies/${id}`)}
        label={formatPolicyName(name, stepNum)}
        style={style}
      />
    )
  }
}

@withRouter
export class ScheduleChip extends Component {
  static propTypes = {
    id: p.string.isRequired,
    style: p.object,
    name: p.string.isRequired,
    onDelete: p.func,
  }

  render() {
    const { id, history, name, onDelete, style } = this.props

    return (
      <Chip
        data-cy='schedule-chip'
        avatar={
          <Avatar>
            <ScheduleIcon />
          </Avatar>
        }
        onDelete={onDelete}
        onClick={() => history.push(`/schedules/${id}`)}
        label={name}
        style={style}
      />
    )
  }
}

export class SlackChip extends Component {
  static propTypes = {
    name: p.string.isRequired,
    onDelete: p.func,
    style: p.object,
  }

  render() {
    const { name, onDelete, style } = this.props

    return (
      <Chip
        data-cy='slack-chip'
        avatar={
          <Avatar>
            <SlackBW />
          </Avatar>
        }
        onDelete={onDelete}
        label={name}
        style={style}
      />
    )
  }
}
