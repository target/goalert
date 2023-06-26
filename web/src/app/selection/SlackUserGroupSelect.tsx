import React, { useEffect, useState } from 'react'
import { gql } from 'urql'
import { makeQuerySelect } from './QuerySelect'
import { SlackChannelSelect } from './SlackChannelSelect'
import { FormControl, FormHelperText } from '@mui/material'

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
  error?: { message: string }
  label?: string
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
    <React.Fragment>
      <FormControl error={Boolean(props.error)}>
        <SlackUserGroupQuerySelect
          value={groupID}
          onChange={handleGroupChange}
          label={props.label}
          name='selectUserGroup'
        />
        <FormHelperText>
          The selected group's membership will be replaced/set to the schedule's
          on-call user(s).
        </FormHelperText>
      </FormControl>

      <FormControl style={{ marginTop: '0.5em' }} error={Boolean(props.error)}>
        <SlackChannelSelect
          label='Error Channel'
          name='errorChannel'
          value={channelID}
          onChange={handleChannelChange}
        />
        <FormHelperText>
          Any problems updating the user group will be sent to this channel.
        </FormHelperText>
      </FormControl>
    </React.Fragment>
  )
}

export default SlackUserGroupSelect
