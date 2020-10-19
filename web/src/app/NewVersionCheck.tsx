import React, { useEffect, useState } from 'react'
import { GOALERT_VERSION, pathPrefix } from './env'
import { DateTime } from 'luxon'
import { Snackbar, Button } from '@material-ui/core'
import {
  UPDATE_CHECK_INTERVAL,
  UPDATE_FORCE_DURATION,
  UPDATE_NOTIF_DURATION,
} from './config'

const fetchCurrentVersion = (): Promise<string> =>
  fetch(pathPrefix)
    .then((res) => res.text())
    .then(
      (docStr) =>
        new DOMParser()
          .parseFromString(docStr, 'text/html')
          .querySelector('meta[http-equiv=x-goalert-version]')
          ?.getAttribute('content') || '',
    )

export default function NewVersionCheck(): JSX.Element {
  const [currentVersion, setCurrentVersion] = useState(GOALERT_VERSION)
  const [firstSeen, setFirstSeen] = useState(DateTime.utc())
  const [lastCheck, setLastCheck] = useState(DateTime.utc())

  useEffect(() => {
    let handleCurrentVersion = (version: string): void => {
      setLastCheck(DateTime.utc())
      if (version === currentVersion) {
        return
      }
      setCurrentVersion(version)
      setFirstSeen(DateTime.utc())
    }
    const ivl = setInterval(() => {
      fetchCurrentVersion().then((version) => handleCurrentVersion(version))
    }, UPDATE_CHECK_INTERVAL.as('millisecond'))

    return () => {
      clearInterval(ivl)
      handleCurrentVersion = () => {}
    }
  }, [currentVersion])

  const hasNewVersion =
    Boolean(currentVersion) && currentVersion !== GOALERT_VERSION

  if (hasNewVersion && lastCheck.diff(firstSeen) >= UPDATE_FORCE_DURATION) {
    // hard-reload after a day
    location.reload()
  }

  return (
    <Snackbar
      open={hasNewVersion && lastCheck.diff(firstSeen) >= UPDATE_NOTIF_DURATION}
      message='A new version is available.'
      anchorOrigin={{
        vertical: 'bottom',
        horizontal: 'left',
      }}
      action={
        <Button
          color='inherit'
          size='small'
          onClick={() => {
            location.reload()
          }}
        >
          Refresh
        </Button>
      }
    />
  )
}
