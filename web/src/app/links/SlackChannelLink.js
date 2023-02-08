import React from 'react'
import AppLink from '../util/AppLink'
import { useQuery, gql } from '@apollo/client'

export const SlackChannelLink = (slackChannel) => {
  const query = gql`
    query ($id: ID!) {
      slackChannel(id: $id) {
        id
        teamID
      }
    }
  `

  const { data, loading, error } = useQuery(query, {
    variables: { id: slackChannel.id },
    fetchPolicy: 'cache-first',
  })
  const teamID = data?.slackChannel?.teamID

  if (error) {
    console.error(`Error querying slackChannel ${slackChannel.id}:`, error)
  }
  if (data && !teamID && !loading) {
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
