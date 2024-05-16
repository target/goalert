import React, { useState } from 'react'
import { gql, useQuery } from 'urql'
import { Redirect } from 'wouter'
import { GenericError, ObjectNotFound } from '../../error-pages'
import { IntegrationKey, Service } from '../../../schema'
import UniversalKeyRuleList from './UniversalKeyRuleList'
import DetailsPage from '../../details/DetailsPage'
import { Action } from '../../details/CardActions'
import GenTokenDialog from './GenTokenDialog'
import PromoteTokenDialog from './PromoteTokenDialog'

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
  if (
    q.data &&
    q.data.integrationKey &&
    q.data.integrationKey.serviceID !== props.serviceID
  ) {
    return (
      <Redirect
        to={`/services/${q.data.integrationKey.serviceID}/integration-keys/${props.keyID}`}
      />
    )
  }

  if (q.error) {
    return <GenericError error={q.error.message} />
  }
  if (!q.data) return <ObjectNotFound type='integration key' />

  const primaryHint = q.data.integrationKey.tokenInfo.primaryHint
  const secondaryHint = q.data.integrationKey.tokenInfo.secondaryHint

  const desc = `
  Primary Token: ${primaryHint || 'N/A'}
  ${secondaryHint ? `\nSecondary Token: ${secondaryHint}` : ''}
  `

  function makeGenerateButtons(): Array<Action> {
    const handleOnClick = (): void => {
      setGenDialogOpen(true)
    }

    if (primaryHint && !secondaryHint) {
      return [
        {
          label: 'Regenerate Token',
          handleOnClick,
        },
      ]
    }

    if (primaryHint && secondaryHint) {
      return [
        {
          label: 'Regenerate Token',
          handleOnClick,
        },
        {
          label: 'Promote Secondary',
          handleOnClick: () => setPromoteDialogOpen(true),
        },
      ]
    }

    return [
      {
        label: 'Generate Token',
        handleOnClick,
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
        pageContent={<UniversalKeyRuleList />}
      />
      <GenTokenDialog
        keyID={props.keyID}
        open={genDialogOpen}
        onClose={() => setGenDialogOpen(false)}
        // isSecondary={}
      />
      <PromoteTokenDialog
        keyID={props.keyID}
        open={promoteDialogOpen}
        onClose={() => setPromoteDialogOpen(false)}
      />
    </React.Fragment>
  )
}
