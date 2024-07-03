import React from 'react'
import { DestinationDisplayInfo, DestinationInput } from '../../schema'
import { gql, useQuery } from 'urql'
import DestinationChip from './DestinationChip'

export type DestinationInputChipProps = {
  value: DestinationInput
  onDelete?: () => void
  onChipClick?: () => void
}

const query = gql`
  query DestDisplayInfo($input: DestinationInput!) {
    destinationDisplayInfo(input: $input) {
      text
      iconURL
      iconAltText
      linkURL
    }
  }
`

const context = {
  suspense: false,
}

// This is a simple wrapper around DestinationChip that takes a DestinationInput
// instead of a DestinationDisplayInfo. It's useful for showing the destination
// chips in the policy details page.
export default function DestinationInputChip(
  props: DestinationInputChipProps,
): React.ReactNode {
  const [{ data, error }] = useQuery<{
    destinationDisplayInfo: DestinationDisplayInfo
  }>({
    query,
    variables: {
      input: props.value,
    },
    requestPolicy: 'cache-first',
    context,
  })

  if (error) {
    return <DestinationChip error={error.message} onDelete={props.onDelete} />
  }

  return (
    <DestinationChip
      iconAltText={data?.destinationDisplayInfo.iconAltText || ''}
      iconURL={data?.destinationDisplayInfo.iconURL || ''}
      linkURL={data?.destinationDisplayInfo.linkURL || ''}
      text={data?.destinationDisplayInfo.text || ''}
      onDelete={props.onDelete}
      onChipClick={props.onChipClick}
    />
  )
}
