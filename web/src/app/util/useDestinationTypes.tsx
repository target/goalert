import React from 'react'
import { gql, useQuery } from 'urql'
import { DestinationType, DestinationTypeInfo } from '../../schema'

const DestTypeContext = React.createContext([] as DestinationTypeInfo[])
DestTypeContext.displayName = 'DestTypeContext'

const query = gql`
  query DestTypes {
    destinationTypes {
      type
      name
      enabled
      disabledMessage
      userDisclaimer

      isContactMethod
      isEPTarget
      isSchedOnCallNotify

      requiredFields {
        fieldID
        labelSingular
        labelPlural
        iconURL
        iconAltText
        hint
        hintURL
        placeholderText
        prefix
        inputType
        isSearchSelectable
        supportsValidation
      }
    }
  }
`

type DestTypeProviderProps = {
  children: React.ReactNode
}

export function DestTypeProvider(
  props: DestTypeProviderProps,
): React.ReactNode {
  const [{ data, error }] = useQuery({ query, requestPolicy: 'cache-first' })
  if (error) throw error

  const destTypes = data?.destinationTypes || []

  return (
    <DestTypeContext.Provider value={destTypes}>
      {props.children}
    </DestTypeContext.Provider>
  )
}

export function useContactMethodTypes(): DestinationTypeInfo[] {
  const destTypes = React.useContext(DestTypeContext)
  return destTypes.filter((t) => t.isContactMethod)
}

export function useEPTargetTypes(): DestinationTypeInfo[] {
  const destTypes = React.useContext(DestTypeContext)
  return destTypes.filter((t) => t.isEPTarget)
}

export function useSchedOnCallNotifyTypes(): DestinationTypeInfo[] {
  const destTypes = React.useContext(DestTypeContext)
  return destTypes.filter((t) => t.isSchedOnCallNotify)
}

export function useDestinationType(type: DestinationType): DestinationTypeInfo {
  const destTypes = React.useContext(DestTypeContext)
  const typeInfo = destTypes.find((t) => t.type === type)

  if (!typeInfo) throw new Error(`unknown destination type '${type}'`)

  return typeInfo
}
