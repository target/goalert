import React, { useEffect, useState } from 'react'
import { GOALERT_VERSION, pathPrefix } from './env'
import { DateTime } from 'luxon'
import { Snackbar, Button } from '@mui/material'
import {
  UPDATE_CHECK_INTERVAL,
  UPDATE_FORCE_DURATION,
  UPDATE_NOTIF_DURATION,
} from './config'

/* extractMetaTagValue extracts the value of a meta tag from an HTML string.
 * It avoids using DOMParser to avoid issues with CSP.
 */
function extractMetaTagValue(htmlString: string, httpEquiv: string): string {
  const lowerHtml = htmlString.toLowerCase()
  const startIndex = lowerHtml.indexOf(
    `<meta http-equiv="${httpEquiv.toLowerCase()}"`,
  )

  if (startIndex === -1) return ''

  const contentStart = lowerHtml.indexOf('content="', startIndex)
  if (contentStart === -1) return ''

  const contentEnd = lowerHtml.indexOf('"', contentStart + 9)
  if (contentEnd === -1) return ''

  return htmlString.slice(contentStart + 9, contentEnd)
}

const fetchCurrentVersion = (): Promise<string> =>
  fetch(pathPrefix)
    .then((res) => res.text())
    .then((docStr) => extractMetaTagValue(docStr, 'x-goalert-version'))

export default function NewVersionCheck(): React.JSX.Element {
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
