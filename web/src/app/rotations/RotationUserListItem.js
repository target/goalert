import React from 'react'
import { PropTypes as p } from 'prop-types'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import { UserAvatar } from '../util/avatars'
import withStyles from '@material-ui/core/styles/withStyles'
import RotationUpdateDialog from './RotationUpdateDialog'
import OtherActions from '../util/OtherActions'

const styles = {
  activeUser: {
    borderLeft: '6px solid #93ed94',
    background: '#defadf',
    width: '100%',
    marginLeft: '0',
    marginRight: '0',
  },
  participantDragging: {
    backgroundColor: '#ebebeb',
  },
}

@withStyles(styles)
export default class RotationUserListItem extends React.PureComponent {
  static propTypes = {
    rotationID: p.string.isRequired,
    userIDs: p.array.isRequired,
    user: p.object,
    index: p.number,
    activeUserIndex: p.number.isRequired,
  }

  state = {
    active: -1,
    delete: false,
  }

  getActiveStyle = index => {
    const { activeUserIndex, classes } = this.props

    // light green left border, lighter green background
    return activeUserIndex === index ? classes.activeUser : null
  }

  render() {
    const { index, user } = this.props
    return (
      <React.Fragment key={user.id}>
        <ListItem key={index} className={this.getActiveStyle(index)}>
          <UserAvatar userID={user.id} />
          <ListItemText
            primary={user.name}
            secondary='Placeholder HandoffTimes Text'
          />
          <OtherActions
            actions={[
              {
                label: 'Set Active',
                onClick: () => this.setState({ active: index }),
              },
              {
                label: 'Remove',
                onClick: () => this.setState({ delete: user.id }),
              },
            ]}
          />
        </ListItem>

        {this.state.active !== -1 && (
          <RotationUpdateDialog
            rotationID={this.props.rotationID}
            onClose={() => this.setState({ active: -1 })}
            activeUserIndex={this.state.active}
            userIDs={this.props.userIDs}
          />
        )}
        {this.state.delete && (
          <RotationUpdateDialog
            rotationID={this.props.rotationID}
            onClose={() => this.setState({ delete: false })}
            userID={this.state.delete}
            userIDs={this.props.userIDs}
          />
        )}
      </React.Fragment>
    )
  }
}
