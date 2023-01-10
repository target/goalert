import React from 'react'
import AppLink from '../util/AppLink'

export const WebhookLink = (webhook: {
  id: string
  name: string
}): JSX.Element => {
  return <AppLink to={webhook.id}>{webhook.name}</AppLink>
}
