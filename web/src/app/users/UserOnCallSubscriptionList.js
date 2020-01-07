import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { Card, Tooltip } from '@material-ui/core'
import FlatList from '../lists/FlatList'
import OtherActions from '../util/OtherActions'
import CreateFAB from '../lists/CreateFAB'
import CalendarSubscribeDialog from '../schedules/calendar-subscribe/CalendarSubscribeDialog'
import { Warning } from '../icons'

export default function UserOnCallSubscriptionList(props) {
  const [showCreateDialog, setShowCreateDialog] = useState(false)

  function renderOtherActions() {
    return (
      <OtherActions
        actions={[
          {
            label: 'Edit',
            onClick: () => {},
          },
          {
            label: 'Delete',
            onClick: () => {},
          },
        ]}
      />
    )
  }

  return (
    <React.Fragment>
      <Card>
        <FlatList
          headerNote='Showing your current on-call subscriptions for all schedules'
          emptyMessage='Your are not subscribed to any schedules'
          items={[
            {
              title: 'My Outlook Calendar',
              subText: (
                <React.Fragment>
                  Central Business Schedule
                  <br />
                  Last access: 48m ago
                </React.Fragment>
              ),
              secondaryAction: renderOtherActions(),
            },
            {
              title: 'My iPhone Calendar',
              icon: (
                <Tooltip title='Disabled'>
                  <Warning />
                </Tooltip>
              ),
              subText: (
                <React.Fragment>
                  Target Main Schedule
                  <br />
                  Last access: 3h ago
                </React.Fragment>
              ),
              secondaryAction: renderOtherActions(),
            },
          ]}
        />
      </Card>
      <CreateFAB
        title='Create Subscription'
        onClick={() => setShowCreateDialog(true)}
      />
      <CalendarSubscribeDialog
        open={showCreateDialog}
        onClose={() => setShowCreateDialog(false)}
      />
    </React.Fragment>
  )
}

UserOnCallSubscriptionList.propTypes = {
  userID: p.string.isRequired,
}
