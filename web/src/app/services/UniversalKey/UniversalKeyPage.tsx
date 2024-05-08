import React from 'react'

interface UniversalKeyPageProps {
  serviceID: string
  keyName: string
}

export default function UniversalKeyPage({
  serviceID,
  keyName,
}: UniversalKeyPageProps): JSX.Element {
  if (serviceID && keyName) {
    return <div />
  }
  return <div />
}
