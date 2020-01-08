import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { Card, Tooltip } from '@material-ui/core'
import FlatList from '../lists/FlatList'
import OtherActions from '../util/OtherActions'
import CreateFAB from '../lists/CreateFAB'
import CalendarSubscribeCreateDialog from '../schedules/calendar-subscribe/CalendarSubscribeCreateDialog'
import { Warning } from '../icons'
import CalendarSubscribeDeleteDialog from '../schedules/calendar-subscribe/CalendarSubscribeDeleteDialog'
import CalendarSubscribeEditDialog from '../schedules/calendar-subscribe/CalendarSubscribeEditDialog'

export default function UserCalendarSubscriptionList(props) {
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showEditDialogByID, setShowEditDialogByID] = useState(null)
  const [showDeleteDialogByID, setShowDeleteDialogByID] = useState(null)

  // todo: query for data here instead
  const data = {
    calendarSubscriptions: mockItems,
  }

  const subs = data.calendarSubscriptions.sort((a, b) => {
    if (a.schedule.name < b.schedule.name) return -1
    if (a.schedule.name > b.schedule.name) return 1

    if (a.name > b.name) return 1
    if (a.name < b.name) return -1
  })

  const subheaderDict = {}
  const items = []

  subs.forEach(sub => {
    if (!subheaderDict[sub.schedule.name]) {
      subheaderDict[sub.schedule.name] = true
      items.push({ subHeader: sub.schedule.name })
    }

    items.push({
      title: sub.name,
      subText: 'Last sync: ' + sub.last_access, // todo: format iso timestamp to duration with luxon
      secondaryAction: renderOtherActions(sub.id),
      icon: sub.disabled ? (
        <Tooltip title='Disabled'>
          <Warning />
        </Tooltip>
      ) : null,
    })
  })

  // todo: finish these dialogs
  function renderOtherActions(id) {
    return (
      <OtherActions
        actions={[
          {
            label: 'Edit',
            onClick: () => setShowEditDialogByID(id),
          },
          {
            label: 'Delete',
            onClick: () => setShowDeleteDialogByID(id),
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
          items={items}
        />
      </Card>
      <CreateFAB
        title='Create Subscription'
        onClick={() => setShowCreateDialog(true)}
      />
      {showCreateDialog && (
        <CalendarSubscribeCreateDialog
          onClose={() => setShowCreateDialog(false)}
        />
      )}
      {showEditDialogByID && (
        <CalendarSubscribeEditDialog
          calSubscriptionID={showEditDialogByID}
          onClose={() => setShowEditDialogByID(null)}
        />
      )}
      {showDeleteDialogByID && (
        <CalendarSubscribeDeleteDialog
          calSubscriptionID={showDeleteDialogByID}
          onClose={() => setShowDeleteDialogByID(null)}
        />
      )}
    </React.Fragment>
  )
}

UserCalendarSubscriptionList.propTypes = {
  userID: p.string.isRequired,
}

const mockItems = [
  {
    id: '1234',
    name: 'asdasd (1)',
    schedule: {
      name: 'Test Schedule 1',
    },
    last_access: '48m ago',
    disabled: false,
  },
  {
    id: '1234',
    name: 'rhgdgb (2)',
    schedule: {
      name: 'Test Schedule 2',
    },
    last_access: '48m ago',
    disabled: false,
  },
  {
    id: '1234',
    name: 'ersefbdfb (3)',
    schedule: {
      name: 'Test Schedule 3',
    },
    last_access: '48m ago',
    disabled: false,
  },
  {
    id: '1234',
    name: 'gfewsfbd (1)',
    schedule: {
      name: 'Test Schedule 1',
    },
    last_access: '48m ago',
    disabled: false,
  },
  {
    id: '1234',
    name: 'wergfs (4)',
    schedule: {
      name: 'Test Schedule 4',
    },
    last_access: '48m ago',
    disabled: false,
  },
  {
    id: '1234',
    name: 'ertbdrgb (1)',
    schedule: {
      name: 'Test Schedule 1',
    },
    last_access: '48m ago',
    disabled: false,
  },
  {
    id: '1234',
    name: 'wergewrw (3)',
    schedule: {
      name: 'Test Schedule 3',
    },
    last_access: '48m ago',
    disabled: true,
  },
]
