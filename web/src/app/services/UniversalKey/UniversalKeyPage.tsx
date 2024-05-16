import React from 'react'
import UniversalKeyRuleList from './UniversalKeyRuleList'
import { gql, useQuery } from 'urql'
import { GenericError, ObjectNotFound } from '../../error-pages'
import { IntegrationKey, Service } from '../../../schema'
import { Redirect } from 'wouter'
import DetailsPage from '../../details/DetailsPage'
import { Action } from '../../details/CardActions'

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

const desc = `
Primary Token: N/A
Secondary Token: N/A
`

export default function UniversalKeyPage(
  props: UniversalKeyPageProps,
): React.ReactNode {
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

  function makeGenerateButtons(): Array<Action> {
    if (primaryHint && !secondaryHint) {
      return [
        {
          label: 'Regenerate Token',
          handleOnClick: () => {},
        },
      ]
    }

    if (primaryHint && secondaryHint) {
      return [
        {
          label: 'Regenerate Token',
          handleOnClick: () => {},
        },
        {
          label: 'Promote Secondary',
          handleOnClick: () => {},
        },
      ]
    }

    return [
      {
        label: 'Generate Token',
        handleOnClick: () => {},
      },
    ]
  }

  return (
    <DetailsPage
      title={q.data.integrationKey.name}
      subheader={`Service: ${q.data.service.name}`}
      details={desc}
      primaryActions={makeGenerateButtons()}
      pageContent={UniversalKeyRuleList()}
    />
  )
}
