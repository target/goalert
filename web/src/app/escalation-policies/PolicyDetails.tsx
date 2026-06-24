import React, { Suspense, useState } from 'react'
import { useQuery, gql } from 'urql'
import _ from 'lodash'
import { Edit, Delete } from '@mui/icons-material'
import PolicyStepsQuery from './PolicyStepsQuery'
import PolicyDeleteDialog from './PolicyDeleteDialog'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'
import DetailsPage from '../details/DetailsPage'
import PolicyEditDialog from './PolicyEditDialog'
import { GenericError, ObjectNotFound } from '../error-pages'
import Spinner from '../loading/components/Spinner'
import { EPAvatar } from '../util/avatars'
import { Redirect } from 'wouter'

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
    }
  }
`

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
