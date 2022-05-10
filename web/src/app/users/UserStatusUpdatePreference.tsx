import React, { ReactElement } from 'react'
import { gql, QueryResult } from '@apollo/client'
import Query from '../util/Query'
import { Mutation } from '@apollo/client/react/components'
import UserContactMethodSelect from './UserContactMethodSelect'
import { User } from '../../schema'

interface MutationInput {
  variables: MutationVariables
}

interface MutationVariables {
  id: string
  cmID: string
}

const query = gql`
  query statusUpdate($id: ID!) {
    user(id: $id) {
      id
      statusUpdateContactMethodID
    }
  }
`
const mutation = gql`
  mutation ($id: ID!, $cmID: ID!) {
    updateUser(input: { id: $id, statusUpdateContactMethodID: $cmID })
  }
`

const disableVal = 'disable'

export default function UserStatusUpdatePreference(props: {
  userID: string
}): JSX.Element {
  function renderControl(
    cmID: string,
    updateCM: (e: React.ChangeEvent<HTMLInputElement>) => void,
  ): ReactElement {
    return (
      <UserContactMethodSelect
        userID={props.userID}
        label='Alert Status Updates'
        helperText='Update me when my alerts are acknowledged or closed'
        name='alert-status-contact-method'
        value={cmID || disableVal}
        onChange={updateCM}
        extraItems={[{ label: 'Disabled', value: disableVal }]}
      />
    )
  }

  function renderMutation(user: User): ReactElement {
    const setCM =
      (commit: (input: MutationInput) => void) =>
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const cmID = e.target.value === disableVal ? '' : e.target.value
        commit({
          variables: {
            id: props.userID,
            cmID,
          },
        })
      }
    return (
      <Mutation mutation={mutation}>
        {(commit: (input: MutationInput) => void) =>
          renderControl(user.statusUpdateContactMethodID, setCM(commit))
        }
      </Mutation>
    )
  }

  return (
    <Query
      query={query}
      variables={{ id: props.userID }}
      render={({ data }: QueryResult) => renderMutation(data.user)}
    />
  )
}
