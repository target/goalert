import { gql, useQuery } from 'urql'

const query = gql`
  query useExpFlag {
    experimentalFlags
  }
`

// useExpFlag is a hook that returns a boolean indicating whether the
// given experimental flag is enabled.
export function useExpFlag(expFlag: ExpFlag): boolean {
  const [{ data }] = useQuery({ query })

  const flags: Array<ExpFlag> = data?.experimentalFlags ?? []

  return flags.includes(expFlag)
}
