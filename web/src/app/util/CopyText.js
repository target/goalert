import React, { useState } from 'react'
import { makeStyles } from '@material-ui/core/styles'
import p from 'prop-types'
import copyToClipboard from './copyToClipboard'
import ContentCopy from 'mdi-material-ui/ContentCopy'
import Tooltip from '@material-ui/core/Tooltip'
import { AppLink } from './AppLink'

const useStyles = makeStyles({
  copyContainer: {
    alignItems: 'center',
    display: 'flex',
    width: 'fit-content',
  },
  icon: {
    paddingRight: 4,
  },
})

export default function CopyText(props) {
  const classes = useStyles()
  const [showTooltip, setShowTooltip] = useState(false)

  return (
    <Tooltip
      onClose={() => setShowTooltip(false)}
      open={showTooltip}
      title='Copied!'
      placement={props.placement || 'right'}
    >
      <AppLink
        className={classes.copyContainer}
        to={props.value}
        onClick={e => {
          const tgt = e.currentTarget.href

          e.preventDefault()
          copyToClipboard(tgt.replace(/^mailto:/, ''))
          setShowTooltip(true)
        }}
      >
        <ContentCopy className={classes.icon} fontSize='small' />
        {props.title}
      </AppLink>
    </Tooltip>
  )
}

CopyText.propTypes = {
  placement: p.string,
  title: p.string,
  value: p.string.isRequired,
}
