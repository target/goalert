import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Alert, Button, Snackbar } from '@mui/material'

function getVapidKey(): Uint8Array | undefined {
  try {
    const meta = document.querySelector('meta[name="vapid-public-key"]') as HTMLMetaElement | null
    let k = (window as any).vapidPublicKey || (window as any).VAPID_PUBLIC_KEY || (meta && meta.content) || ''
    if (!k) return undefined
    k = k.replace(/-/g, '+').replace(/_/g, '/')
    const pad = '='.repeat((4 - (k.length % 4)) % 4)
    const raw = atob(k + pad)
    const out = new Uint8Array(raw.length)
    for (let i = 0; i < raw.length; i++) out[i] = raw.charCodeAt(i)
    return out
  } catch {
    return undefined
  }
}

async function ensureSubscribed(): Promise<void> {
  if (!('serviceWorker' in navigator) || !(window as any).Notification) return
  const perm = (window as any).Notification.permission
  if (perm !== 'granted') return
  const reg = await navigator.serviceWorker.ready
  let sub = await reg.pushManager.getSubscription()
  if (!sub) {
    sub = await reg.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: getVapidKey() })
  }
  await fetch((window as any).pathPrefix + '/api/push/subscribe', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(sub),
    credentials: 'same-origin',
  })
}

const DISMISS_KEY = 'push-banner-dismissed'

export default function PushPermissionBanner(): JSX.Element | null {
  const supported = typeof window !== 'undefined' && 'serviceWorker' in navigator && (window as any).Notification
  const [open, setOpen] = useState(false)

  const shouldShow = useMemo(() => {
    if (!supported) return false
    const perm = (window as any).Notification.permission
    if (perm === 'granted') return false
    if (localStorage.getItem(DISMISS_KEY) === '1') return false
    return perm === 'default'
  }, [supported])

  useEffect(() => {
    setOpen(shouldShow)
  }, [shouldShow])

  useEffect(() => {
    // if already granted, make sure subscription is stored
    ensureSubscribed().catch((e) => console.error('[push] ensureSubscribed failed', e))
  }, [])

  const onDismiss = useCallback(() => {
    try { localStorage.setItem(DISMISS_KEY, '1') } catch {}
    setOpen(false)
  }, [])

  const onEnable = useCallback(() => {
    try {
      ;(window as any).Notification.requestPermission().then((permission: string) => {
        console.log('[push] permission response (banner):', permission)
        if (permission === 'granted') {
          ensureSubscribed().catch((e) => console.error('[push] subscribe/store failed', e))
          setOpen(false)
        }
        if (permission === 'denied') {
          setOpen(false)
        }
      })
    } catch (e) {
      console.error('[push] permission request threw', e)
    }
  }, [])

  if (!open) return null

  return (
    <Snackbar open anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}>
      <Alert elevation={6} variant='filled' severity='info'
        action={
          <>
            <Button color='inherit' size='small' onClick={onEnable}>
              Allow
            </Button>
            <Button color='inherit' size='small' onClick={onDismiss}>
              Dismiss
            </Button>
          </>
        }
      >
        Enable push notifications?
      </Alert>
    </Snackbar>
  )
}

