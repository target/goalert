import React from 'react'
import BaseAvatar from './BaseAvatar'
import { Layers, RotateRight, Today, VpnKey, Person } from '@material-ui/icons'

import { graphql } from 'react-apollo'
import gql from 'graphql-tag'

export class UserAvatar extends BaseAvatar {
  renderFallback() {
    return <Person />
  }
  srcURL(props) {
    return props.userID && `/api/v2/user-avatar/${props.userID}`
  }
}
export class ServiceAvatar extends BaseAvatar {
  renderFallback() {
    return <VpnKey />
  }
}
export class EPAvatar extends BaseAvatar {
  renderFallback() {
    return <Layers />
  }
}
export class RotationAvatar extends BaseAvatar {
  renderFallback() {
    return <RotateRight />
  }
}
export class ScheduleAvatar extends BaseAvatar {
  renderFallback() {
    return <Today />
  }
}

const CURRENT_USER_QUERY = gql`
  query GetCurrentUser {
    currentUser {
      id
    }
  }
`

@graphql(CURRENT_USER_QUERY)
export class CurrentUserAvatar extends UserAvatar {
  render() {
    const { data, loading, error, ...props } = this.props
    const userID = data && data.currentUser && data.currentUser.id

    return <UserAvatar {...props} userID={userID} />
  }
}
