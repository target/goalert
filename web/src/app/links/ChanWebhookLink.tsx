import React from 'react'
import AppLink from '../util/AppLink'

export const ChanWebhookLink = (chanWebhook: {
  id: string
  name: string
}): JSX.Element => {
  return <AppLink to={chanWebhook.id}>{chanWebhook.name}</AppLink>
}
