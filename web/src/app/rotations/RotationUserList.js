import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import FlatList from '../lists/FlatList'
import Query from '../util/Query'
import Card from '@material-ui/core/Card'
import CardHeader from '@material-ui/core/CardHeader'
import { reorderList, calcNewActiveIndex } from './util'
import { Mutation } from 'react-apollo'
import OtherActions from '../util/OtherActions'
import CountDown from '../util/CountDown'
import RotationSetActiveDialog from './RotationSetActiveDialog'
import RotationUserDeleteDialog from './RotationUserDeleteDialog'
import { DateTime } from 'luxon'
import { UserAvatar } from '../util/avatar'
import { withStyles } from '@material-ui/core'
import { styles as globalStyles } from '../styles/materialStyles'

const rotationUsersQuery = gql`
  query rotationUsers($id: ID!) {
    rotation(id: $id) {
      id
      users {
        id
        name
      }
      activeUserIndex
      nextHandoffTimes
    }
  }
`

const mutation = gql`
  mutation updateRotation($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`

const styles = theme => {
  const { cardHeader } = globalStyles(theme)

  return {
    cardHeader,
  }
}

@withStyles(styles)
export default class RotationUserList extends React.PureComponent {
  static propTypes = {
    rotationID: p.string.isRequired,
  }

  state = {
    deleteIndex: null,
    setActiveIndex: null,
  }

  render() {
    const { classes } = this.props
    return (
      <React.Fragment>
        <Card>
          <CardHeader
            className={classes.cardHeader}
            component='h3'
            title='Users'
          />
          <Query
            query={rotationUsersQuery}
            render={({ data }) => this.renderMutation(data)}
            variables={{ id: this.props.rotationID }}
          />
        </Card>
        {this.state.deleteIndex !== null && (
          <RotationUserDeleteDialog
            rotationID={this.props.rotationID}
            userIndex={this.state.deleteIndex}
            onClose={() => this.setState({ deleteIndex: null })}
          />
        )}
        {this.state.setActiveIndex !== null && (
          <RotationSetActiveDialog
            rotationID={this.props.rotationID}
            userIndex={this.state.setActiveIndex}
            onClose={() => this.setState({ setActiveIndex: null })}
          />
        )}
      </React.Fragment>
    )
  }

  renderMutation(data) {
    return (
      <Mutation mutation={mutation}>
        {commit => this.renderList(data, commit)}
      </Mutation>
    )
  }

  renderList(data, commit) {
    const { users, activeUserIndex, nextHandoffTimes } = data.rotation

    // duplicate first entry
    const _nextHandoffTimes = (nextHandoffTimes || [])
      .slice(0, 1)
      .concat(nextHandoffTimes)
    const handoff = users.map((u, index) => {
      const handoffIndex =
        (index + (users.length - activeUserIndex)) % users.length
      const time = _nextHandoffTimes[handoffIndex]
      if (!time) {
        return null
      }

      if (index === activeUserIndex) {
        return (
          <CountDown
            end={time}
            weeks
            days
            hours
            minutes
            prefix='Active for the next '
            style={{ marginLeft: '1em' }}
            expiredTimeout={60}
            expiredMessage='< 1 Minute'
          />
        )
      }
      return (
        'Starts at ' +
        DateTime.fromISO(time).toLocaleString(DateTime.TIME_SIMPLE) +
        ' ' +
        DateTime.fromISO(time).toRelativeCalendar()
      )
    })

    return (
      <FlatList
        data-cy='users'
        emptyMessage='No users currently assigned to this rotation'
        headerNote={
          users.length ? "Click and drag on a user's name to re-order" : ''
        }
        items={users.map((u, index) => ({
          title: u.name,
          id: u.id,
          highlight: index === activeUserIndex,
          icon: <UserAvatar userID={u.id} />,
          subText: handoff[index],
          secondaryAction: (
            <OtherActions
              actions={[
                {
                  label: 'Set Active',
                  onClick: () => this.setState({ setActiveIndex: index }),
                },
                {
                  label: 'Remove',
                  onClick: () => this.setState({ deleteIndex: index }),
                },
              ]}
            />
          ),
        }))}
        onReorder={(...args) => {
          let updatedUsers = reorderList(users.map(u => u.id), ...args)
          const newActiveIndex = calcNewActiveIndex(activeUserIndex, ...args)
          const params = { id: this.props.rotationID, userIDs: updatedUsers }

          if (newActiveIndex !== -1) {
            params.activeUserIndex = newActiveIndex
          }

          return commit({
            variables: { input: params },
            update: (cache, response) => {
              if (!response.data.updateRotation) {
                return
              }
              const data = cache.readQuery({
                query: rotationUsersQuery,
                variables: { id: this.props.rotationID },
              })

              const users = reorderList(data.rotation.users, ...args)

              cache.writeQuery({
                query: rotationUsersQuery,
                variables: { id: this.props.rotationID },
                data: {
                  ...data,
                  rotation: {
                    ...data.rotation,
                    activeUserIndex:
                      newActiveIndex === -1
                        ? data.rotation.activeUserIndex
                        : newActiveIndex,
                    users,
                  },
                },
              })
            },
            optimisticResponse: {
              __typename: 'Mutation',
              updateRotation: true,
            },
          })
        }}
      />
    )
  }
}
