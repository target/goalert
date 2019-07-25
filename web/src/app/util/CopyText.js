import React, { useState } from 'react'
import p from 'prop-types'
import copyToClipboard from './copyToClipboard'
import ContentCopy from 'mdi-material-ui/ContentCopy'
import Tooltip from '@material-ui/core/Tooltip'

export default function CopyText(props) {
  const [showTooltip, setShowTooltip] = useState(false)

  return (
    <Tooltip
      onClose={() => setShowTooltip(false)}
      open={showTooltip}
      title='Copied!'
      placement='right'
    >
      <a
        href={props.value}
        onClick={e => {
          e.preventDefault()
          copyToClipboard(props.value)
          setShowTooltip(true)
        }}
      >
        <ContentCopy />
        {props.title}
      </a>
    </Tooltip>
  )
}

CopyText.propTypes = {
  title: p.string.isRequired,
  value: p.string.isRequired,
}
