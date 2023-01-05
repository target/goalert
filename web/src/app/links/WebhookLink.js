import React from 'react'
import AppLink from '../util/AppLink'

export const WebhookLink = (webhook) => {
  return <AppLink to={webhook.id}>{webhook.name}</AppLink>
}
