import React, { useState } from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import DetailsPage from '../details/DetailsPage'
import StatusUpdateNotification from './UserStatusUpdatePreference'
import { UserAvatar } from '../util/avatar'
import UserContactMethodList from './UserContactMethodList'
import { AddAlarm, SettingsPhone } from '@material-ui/icons'
import SpeedDial from '../util/SpeedDial'
import UserNotificationRuleList from './UserNotificationRuleList'
import { Grid } from '@material-ui/core'
import UserContactMethodCreateDialog from './UserContactMethodCreateDialog'
import UserNotificationRuleCreateDialog from './UserNotificationRuleCreateDialog'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core/styles'
import UserContactMethodVerificationDialog from './UserContactMethodVerificationDialog'
import { useQuery } from '@apollo/react-hooks'
import _ from 'lodash-es'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'
import { useConfigValue } from '../util/RequireConfig'

const query = gql`
  query userInfo($id: ID!) {
    user(id: $id) {
      id
      name
      email
      onCallSteps {
        id
        escalationPolicy {
          id
          assignedTo {
            id
            name
          }
        }
      }
    }
  }
`

const useStyles = makeStyles({
  gravatarText: {
    textAlign: 'center',
    paddingTop: '0.5em',
    display: 'block',
  },
  profileImage: {
    width: 128,
    height: 128,
    margin: 'auto',
  },
})

function serviceCount(onCallSteps = []) {
  const svcs = {}
  ;(onCallSteps || []).forEach(s =>
    (s.escalationPolicy.assignedTo || []).forEach(svc => (svcs[svc.id] = true)),
  )

  return Object.keys(svcs).length
}

export default function UserDetails(props) {
  const classes = useStyles()

  const [disclaimer] = useConfigValue('General.NotificationDisclaimer')
  const [createCM, setCreateCM] = useState(false)
  const [createNR, setCreateNR] = useState(false)
  const [showVerifyDialogByID, setShowVerifyDialogByID] = useState(null)

  const { data, loading, error } = useQuery(query, {
    variables: { id: props.userID },
  })

  if (error) return <GenericError error={error.message} />
  if (!_.get(data, 'user.id')) return loading ? <Spinner /> : <ObjectNotFound />

  const user = _.get(data, 'user')
  const svcCount = serviceCount(user.onCallSteps)

  return (
    <React.Fragment>
      {props.readOnly ? null : (
        <SpeedDial
          label='Add Items'
          actions={[
            {
              label: 'Add Contact Method',
              icon: <SettingsPhone />,
              onClick: () => setCreateCM(true),
            },
            {
              label: 'Add Notification Rule',
              icon: <AddAlarm />,
              onClick: () => setCreateNR(true),
            },
          ]}
        />
      )}
      {createCM && (
        <UserContactMethodCreateDialog
          userID={props.userID}
          disclaimer={disclaimer}
          onClose={result => {
            setCreateCM(false)
            setShowVerifyDialogByID(
              result && result.contactMethodID ? result.contactMethodID : null,
            )
          }}
        />
      )}
      {showVerifyDialogByID && (
        <UserContactMethodVerificationDialog
          contactMethodID={showVerifyDialogByID}
          onClose={() => setShowVerifyDialogByID(null)}
        />
      )}
      {createNR && (
        <UserNotificationRuleCreateDialog
          userID={props.userID}
          onClose={() => setCreateNR(false)}
        />
      )}
      <DetailsPage
        title={user.name + (svcCount ? ' (On-Call)' : '')}
        details={user.email}
        icon={
          <React.Fragment>
            <UserAvatar
              userID={props.userID}
              className={classes.profileImage}
            />
            <Typography variant='caption' className={classes.gravatarText}>
              Provided by{' '}
              <a
                href='https://gravatar.com'
                target='_blank'
                rel='noopener noreferrer'
              >
                Gravatar
              </a>
            </Typography>
          </React.Fragment>
        }
        links={[
          {
            label: 'On-Call Assignments',
            url: 'on-call-assignments',
            subText: svcCount
              ? `On-call for ${svcCount} service${svcCount > 1 ? 's' : ''}`
              : 'Not currently on-call',
          },
        ]}
        titleFooter={
          props.readOnly ? null : (
            <StatusUpdateNotification userID={props.userID} />
          )
        }
        pageFooter={
          <Grid container spacing={2}>
            <UserContactMethodList
              userID={props.userID}
              readOnly={props.readOnly}
            />
            <UserNotificationRuleList
              userID={props.userID}
              readOnly={props.readOnly}
            />
          </Grid>
        }
      />
    </React.Fragment>
  )
}

UserDetails.propTypes = {
  userID: p.string.isRequired,
  readOnly: p.bool,
}
