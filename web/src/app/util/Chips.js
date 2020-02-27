import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Chip from '@material-ui/core/Chip'
import { withRouter } from 'react-router-dom'
import { push } from 'connected-react-router'
import { useDispatch } from 'react-redux'
import { useQuery } from 'react-apollo'

import {
  Layers as PolicyIcon,
  RotateRight as RotationIcon,
  Today as ScheduleIcon,
} from '@material-ui/icons'
import Avatar from '@material-ui/core/Avatar'
import { UserAvatar, ServiceAvatar } from './avatar'
import { SlackBW } from '../icons'
import gql from 'graphql-tag'

const serviceQuery = gql`
  query service($id: ID!) {
    service(id: $id) {
      id
      name
    }
  }
`

export function ServiceChip(props) {
  const { className, id, name, onDelete, style, onClick } = props
  const dispatch = useDispatch()

  const { data, loading, error } = useQuery(serviceQuery, {
    variables: {
      id,
    },
    skip: Boolean(name),
    fetchPolicy: 'cache-first',
    pollInterval: 0,
  })

  const getLabel = () => {
    if (name) return name

    if (!data && loading) return 'Loading...'

    if (error) return `Error: ${error.message}`

    return data.service.name
  }

  return (
    <Chip
      className={className}
      data-cy='service-chip'
      avatar={<ServiceAvatar />}
      style={style}
      onDelete={onDelete}
      onClick={onClick || (() => dispatch(push(`/services/${id}`)))}
      label={getLabel()}
    />
  )
}
ServiceChip.propTypes = {
  id: p.string.isRequired,
  style: p.object,
  name: p.string,
  onDelete: p.func,
  onClick: p.func,
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
