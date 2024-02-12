import React from 'react'
import { gql, useQuery } from 'urql'

const query = gql`
  query useExpFlag {
    experimentalFlags
  }
`

// useExpFlag is a hook that returns a boolean indicating whether the
// given experimental flag is enabled.
export function useExpFlag(expFlag: ExpFlag): boolean {
  const [{ data }] = useQuery({ query, requestPolicy: 'cache-first' })

  const flags: Array<ExpFlag> = data?.experimentalFlags ?? []

  return flags.includes(expFlag)
}

// ExpFlag is used to conditionally render experimental features.
//
// Example:
//
//   <ExpFlag flag="my-flag">
//     <MyFeature />
//   </ExpFlag>
//
export function ExpFlag(props: {
  flag: ExpFlag // The flag that must be enabled for the children to be rendered.
  children: React.ReactNode
}): React.ReactNode | null {
  const enabled = useExpFlag(props.flag)

  if (!enabled) return null

  return props.children
}
