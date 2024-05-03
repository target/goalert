import React from 'react'

interface UniversalKeyPageProps {
  serviceID: string
  keyID: string
}

export default function UniversalKeyPage({
  serviceID,
  keyID,
}: UniversalKeyPageProps): JSX.Element {
  if (serviceID && keyID) {
    return <div />
  }
  return <div />
}
