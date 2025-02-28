import React from 'react'
import AppLink from '../util/AppLink'
import { useQuery, gql } from 'urql'
import { Target } from '../../schema'

const query = gql`
  query ($id: ID!) {
    slackChannel(id: $id) {
      id
      teamID
    }
  }
`

export const SlackChannelLink = (slackChannel: Target): React.JSX.Element => {
  const [{ data, fetching, error }] = useQuery({
    query,
    variables: { id: slackChannel.id },
    requestPolicy: 'cache-first',
  })
  const teamID = data?.slackChannel?.teamID

  if (error) {
    console.error(`Error querying slackChannel ${slackChannel.id}:`, error)
  }
  if (data && !teamID && !fetching) {
    console.error('Error generating Slack link: team ID not found')
  }

  return (
    <AppLink
      to={`https://slack.com/app_redirect?channel=${slackChannel.id}&team=${teamID}`}
    >
      {slackChannel.name}
    </AppLink>
  )
}
