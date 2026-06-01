import React, { Suspense, useState } from 'react'
import { useQuery, gql } from 'urql'
import _ from 'lodash'
import { Edit, Delete } from '@mui/icons-material'
import PolicyStepsQuery from './PolicyStepsQuery'
import PolicyDeleteDialog from './PolicyDeleteDialog'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'
import DetailsPage, { LinkStatus } from '../details/DetailsPage'
import PolicyEditDialog from './PolicyEditDialog'
import { GenericError, ObjectNotFound } from '../error-pages'
import Spinner from '../loading/components/Spinner'
import { EPAvatar } from '../util/avatars'
import { Redirect } from 'wouter'
import { Target } from 'web/src/schema'

const query = gql`
  query ($id: ID!) {
    escalationPolicy(id: $id) {
      id
      name
      description

      notices {
        type
        message
        details
      }

      assignedTo {
        id
        type
      }
    }
  }
`
const alertStatusQuery = gql`
  query policyAlertStatusQuery($serviceIDs: [ID!]!) {
    alerts(
      input: {
        filterByStatus: [StatusAcknowledged, StatusUnacknowledged]
        filterByServiceID: $serviceIDs
        first: 1
      }
    ) {
      nodes {
        id
        status
      }
    }
  }
`

const alertStatus = (a: { status: string }[]): LinkStatus | null => {
  if (!a) return null
  if (!a.length) return 'ok'
  if (a[0].status === 'StatusUnacknowledged') return 'err'
  return 'warn'
}

export default function PolicyDetails(props: {
  policyID: string
}): React.JSX.Element {
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [showEditDialog, setShowEditDialog] = useState(false)

  const [{ fetching, error, data: _data }] = useQuery({
    query,
    variables: {
      id: props.policyID,
    },
  })
  const [alertStatusResult] = useQuery({
    query: alertStatusQuery,
    variables: {
      serviceIDs: _data?.escalationPolicy?.assignedTo
        .filter((a: Target) => a.type === 'service')
        .map((a: Target) => a.id),
    },
  })

  const data = _.get(_data, 'escalationPolicy', null)

  if (!data && fetching) return <Spinner />
  if (error) return <GenericError error={error.message} />

  if (!data) {
    return showDeleteDialog ? (
      <Redirect to='/escalation-policies' />
    ) : (
      <ObjectNotFound />
    )
  }

  return (
    <React.Fragment>
      <DetailsPage
        notices={data.notices}
        avatar={<EPAvatar />}
        title={data.name}
        details={data.description}
        pageContent={<PolicyStepsQuery escalationPolicyID={data.id} />}
        secondaryActions={[
          {
            label: 'Edit',
            icon: <Edit />,
            handleOnClick: () => setShowEditDialog(true),
          },
          {
            label: 'Delete',
            icon: <Delete />,
            handleOnClick: () => setShowDeleteDialog(true),
          },
          <QuerySetFavoriteButton
            key='secondary-action-favorite'
            id={data.id}
            type='escalationPolicy'
          />,
        ]}
        links={[
          {
            label: 'Alerts',
            status: alertStatus(alertStatusResult.data?.alerts?.nodes),
            url: 'alerts',
            subText: 'Manage alerts specific to services using this policy',
          },
          {
            label: 'Services',
            url: 'services',
            subText: 'Find services that link to this policy',
          },
        ]}
      />
      <Suspense>
        {showEditDialog && (
          <PolicyEditDialog
            escalationPolicyID={data.id}
            onClose={() => setShowEditDialog(false)}
          />
        )}
        {showDeleteDialog && (
          <PolicyDeleteDialog
            escalationPolicyID={data.id}
            onClose={() => setShowDeleteDialog(false)}
          />
        )}
      </Suspense>
    </React.Fragment>
  )
}
