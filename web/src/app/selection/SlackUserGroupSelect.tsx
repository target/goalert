import { gql } from '@apollo/client'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query ($input: SlackUserGroupSearchOptions) {
    slackUserGroups(input: $input) {
      nodes {
        id
        name
        handle
      }
    }
  }
`

const valueQuery = gql`
  query ($id: ID!) {
    slackUserGroup(id: $id) {
      id
      name
      handle
    }
  }
`

export const SlackUserGroupSelect = makeQuerySelect('SlackUserGroupSelect', {
  query,
  valueQuery,
})
