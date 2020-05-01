import React, { useState } from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { useQuery } from '@apollo/react-hooks'
import { Redirect } from 'react-router-dom'
import _ from 'lodash-es'

import PageActions from '../util/PageActions'
import PolicyStepsQuery from './PolicyStepsQuery'
import OtherActions from '../util/OtherActions'
import PolicyDeleteDialog from './PolicyDeleteDialog'
import CreateFAB from '../lists/CreateFAB'
import PolicyStepCreateDialog from './PolicyStepCreateDialog'
import DetailsPage from '../details/DetailsPage'
import PolicyEditDialog from './PolicyEditDialog'
import { useResetURLParams, useURLParam } from '../actions'
import { GenericError, ObjectNotFound } from '../error-pages'
import Spinner from '../loading/components/Spinner'

const query = gql`
  query($id: ID!) {
    escalationPolicy(id: $id) {
      id
      name
      description
    }
  }
`

export default function PolicyDetails(props) {
  const stepNumParam = 'createStep'
  const [createStep, setCreateStep] = useURLParam(stepNumParam, false)
  const resetCreateStep = useResetURLParams(stepNumParam)

  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [showEditDialog, setShowEditDialog] = useState(false)

  const { loading, error, data: _data } = useQuery(query, {
    variables: {
      id: props.escalationPolicyID,
    },
  })

  const data = _.get(_data, 'escalationPolicy', null)

  if (!data && loading) return <Spinner />
  if (error) return <GenericError error={error.message} />

  if (!data) {
    return showDeleteDialog ? (
      <Redirect to='/escalation-policies' push />
    ) : (
      <ObjectNotFound />
    )
  }

  return (
    <React.Fragment>
      <PageActions>
        <OtherActions
          actions={[
            {
              label: 'Edit Escalation Policy',
              onClick: () => setShowEditDialog(true),
            },
            {
              label: 'Delete Escalation Policy',
              onClick: () => setShowDeleteDialog(true),
            },
          ]}
        />
      </PageActions>
      <DetailsPage
        title={data.name}
        details={data.description}
        links={[
          {
            label: 'Services',
            url: 'services',
          },
        ]}
        pageFooter={<PolicyStepsQuery escalationPolicyID={data.id} />}
      />
      <CreateFAB onClick={() => setCreateStep(true)} title='Create Step' />
      {createStep && (
        <PolicyStepCreateDialog
          escalationPolicyID={data.id}
          onClose={resetCreateStep}
        />
      )}
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
    </React.Fragment>
  )
}

PolicyDetails.propTypes = {
  escalationPolicyID: p.string.isRequired,
}
