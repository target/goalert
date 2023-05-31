import React, { useEffect, useState } from 'react'
import { gql } from 'urql'
import { makeQuerySelect } from './QuerySelect'
import { SlackChannelSelect } from './SlackChannelSelect'

const query = gql`
  query ($input: SlackUserGroupSearchOptions) {
    slackUserGroups(input: $input) {
      nodes {
        id
        name: handle
      }
    }
  }
`

const valueQuery = gql`
  query ($id: ID!) {
    slackUserGroup(id: $id) {
      id
      name: handle
    }
  }
`

const SlackUserGroupQuerySelect = makeQuerySelect('SlackUserGroupSelect', {
  query,
  valueQuery,
})

export type SlackUserGroupSelectProps = {
  value: string | null
  onChange: (newValue: string | null) => void
}

export const SlackUserGroupSelect: React.FC<SlackUserGroupSelectProps> = (
  props,
) => {
  const [groupID, setGroupID] = useState<string | null>(null)
  const [channelID, setChannelID] = useState<string | null>(null)

  useEffect(() => {
    if (!props.value) return
    const [groupID, channelID] = props.value?.split(':') || [null, null]
    setGroupID(groupID)
    setChannelID(channelID)
  }, [props.value])

  function handleGroupChange(newGroupID: string | null): void {
    setGroupID(newGroupID)
    if (newGroupID && channelID) {
      props.onChange(`${newGroupID}:${channelID}`)
    }
  }
  function handleChannelChange(newChannelID: string | null): void {
    setChannelID(newChannelID)
    if (newChannelID && groupID) {
      props.onChange(`${groupID}:${newChannelID}`)
    }
  }

  return (
    <div>
      <SlackUserGroupQuerySelect value={groupID} onChange={handleGroupChange} />
      <SlackChannelSelect value={channelID} onChange={handleChannelChange} />
    </div>
  )
}

export default SlackUserGroupSelect
