import React, { useState } from 'react'
import { gql, useQuery } from 'urql'
import { Redirect } from 'wouter'
import { GenericError, ObjectNotFound } from '../../error-pages'
import { IntegrationKey, Service } from '../../../schema'
import DetailsPage from '../../details/DetailsPage'
import { Action } from '../../details/CardActions'
import GenTokenDialog from './GenTokenDialog'
import PromoteTokenDialog from './PromoteTokenDialog'
import DeleteSecondaryTokenDialog from './DeleteSecondaryTokenDialog'

interface UniversalKeyPageProps {
  serviceID: string
  keyID: string
}

const query = gql`
  query UniversalKeyPage($keyID: ID!, $serviceID: ID!) {
    integrationKey(id: $keyID) {
      id
      name
      serviceID
      tokenInfo {
        primaryHint
        secondaryHint
      }
    }
    service(id: $serviceID) {
      id
      name
    }
  }
`

export default function UniversalKeyPage(
  props: UniversalKeyPageProps,
): React.ReactNode {
  const [genDialogOpen, setGenDialogOpen] = useState(false)
  const [promoteDialogOpen, setPromoteDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)

  const [q] = useQuery<{
    integrationKey: IntegrationKey
    service: Service
  }>({
    query,
    variables: {
      keyID: props.keyID,
      serviceID: props.serviceID,
    },
  })

  // Redirect to the correct service if the key is not in the service
  if (q.data?.integrationKey.serviceID !== props.serviceID) {
    return (
      <Redirect
        to={`/services/${q.data?.integrationKey.serviceID}/integration-keys/${props.keyID}`}
      />
    )
  }

  if (q.error) {
    return <GenericError error={q.error.message} />
  }
  if (!q.data) return <ObjectNotFound type='integration key' />

  const primaryHint = q.data.integrationKey.tokenInfo.primaryHint
  const secondaryHint = q.data.integrationKey.tokenInfo.secondaryHint

  const desc = secondaryHint
    ? `Primary Auth Token: ${primaryHint}\nSecondary Auth Token: ${secondaryHint}`
    : `Auth Token: ${primaryHint || 'N/A'}`

  function makeGenerateButtons(): Array<Action> {
    if (primaryHint && !secondaryHint) {
      return [
        {
          label: 'Generate Secondary Token',
          handleOnClick: () => setGenDialogOpen(true),
        },
      ]
    }

    if (primaryHint && secondaryHint) {
      return [
        {
          label: 'Delete Secondary Token',
          handleOnClick: () => setDeleteDialogOpen(true),
        },
        {
          label: 'Promote Secondary Token',
          handleOnClick: () => setPromoteDialogOpen(true),
        },
      ]
    }

    return [
      {
        label: 'Generate Auth Token',
        handleOnClick: () => setGenDialogOpen(true),
      },
    ]
  }

  return (
    <React.Fragment>
      <DetailsPage
        title={q.data.integrationKey.name}
        subheader={`Service: ${q.data.service.name}`}
        details={desc}
        primaryActions={makeGenerateButtons()}
        pageContent={<div />}
      />
      {genDialogOpen && (
        <GenTokenDialog
          keyID={props.keyID}
          onClose={() => setGenDialogOpen(false)}
        />
      )}
      {promoteDialogOpen && (
        <PromoteTokenDialog
          keyID={props.keyID}
          onClose={() => setPromoteDialogOpen(false)}
        />
      )}
      {deleteDialogOpen && (
        <DeleteSecondaryTokenDialog
          keyID={props.keyID}
          onClose={() => setDeleteDialogOpen(false)}
        />
      )}
    </React.Fragment>
  )
}
