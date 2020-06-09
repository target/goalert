import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import { UserAvatar } from '../util/avatars'
import RotationUpdateDialog from './RotationUpdateDialog'
import OtherActions from '../util/OtherActions'

const useStyles = {
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

export default function RotationUserListItem(props) {
  const classes = useStyles()
  const [active, setActive] = useState(-1)
  const [deleteUser, setDeleteUser] = useState(false)

  function getActiveStyle(index) {
    // light green left border, lighter green background
    return props.activeUserIndex === index ? classes.activeUser : null
  }

  return (
    <React.Fragment key={props.user.id}>
      <ListItem key={props.index} className={getActiveStyle(props.index)}>
        <UserAvatar userID={props.user.id} />
        <ListItemText
          primary={props.user.name}
          secondary='Placeholder HandoffTimes Text'
        />
        <OtherActions
          actions={[
            {
              label: 'Set Active',
              onClick: () => setActive(props.index),
            },
            {
              label: 'Remove',
              onClick: () => setDeleteUser(props.user.id),
            },
          ]}
        />
      </ListItem>

      {active !== -1 && (
        <RotationUpdateDialog
          rotationID={props.rotationID}
          onClose={() => setActive(-1)}
          activeUserIndex={active}
          userIDs={props.userIDs}
        />
      )}
      {deleteUser && (
        <RotationUpdateDialog
          rotationID={props.rotationID}
          onClose={() => setDeleteUser(false)}
          userID={deleteUser}
          userIDs={props.userIDs}
        />
      )}
    </React.Fragment>
  )
}

RotationUserListItem.PropTypes = {
  rotationID: p.string.isRequired,
  userIDs: p.array.isRequired,
  user: p.object,
  index: p.number,
  activeUserIndex: p.number.isRequired,
}
