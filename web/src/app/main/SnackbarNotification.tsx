import { Alert, Snackbar } from '@mui/material'
import React, { ReactNode, useState } from 'react'

type Notification = {
  message: string
  severity: 'info' | 'error'
  action?: ReactNode
}

interface NotificationContextParams {
  notification?: Notification
  setNotification: (n: Notification) => void
}

export const NotificationContext =
  React.createContext<NotificationContextParams>({
    notification: undefined,
    setNotification: () => {},
  })
NotificationContext.displayName = 'NotificationContext'

interface NotificationProviderProps {
  children: ReactNode
}

export const NotificationProvider = (
  props: NotificationProviderProps,
): React.ReactNode => {
  const [notification, setNotification] = useState<Notification>()
  const [open, setOpen] = useState(false)

  return (
    <NotificationContext.Provider
      value={{
        notification,
        setNotification: (n) => {
          setNotification(n)
          setOpen(true)
        },
      }}
    >
      {props.children}
      <Snackbar
        action={notification?.action}
        open={open}
        onClose={() => setOpen(false)}
        autoHideDuration={notification?.severity === 'error' ? null : 6000}
        TransitionProps={{ onExited: () => setNotification(undefined) }}
      >
        <Alert
          severity={notification?.severity}
          onClose={() => setOpen(false)}
          variant='filled'
        >
          {notification?.message}
        </Alert>
      </Snackbar>
    </NotificationContext.Provider>
  )
}
