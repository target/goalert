import React, { useState } from 'react'
import { makeStyles } from '@material-ui/core/styles'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogTitle from '@material-ui/core/DialogTitle'
import Divider from '@material-ui/core/Divider'
import { gql, useLazyQuery } from '@apollo/client'
import CardActions from '../details/CardActions'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import Markdown from '../util/Markdown'
import copyToClipboard from '../util/copyToClipboard'

const query = gql`
  query {
    generateSlackAppManifest
  }
`

const useStyles = makeStyles({
  dialog: {
    minHeight: '650px',
  },
})

export default function SlackActions(): JSX.Element {
  const classes = useStyles()
  const [showManifest, setShowManifest] = useState(false)
  const [copied, setCopied] = useState(false)

  const [getManifest, { called, loading, error, data }] = useLazyQuery(query, {
    pollInterval: 0,
  })

  function renderContent(): JSX.Element {
    if (called && loading) return <Spinner />
    if (error) return <GenericError error={error.message} />
    return <Markdown value={'```' + data?.generateSlackAppManifest + '\n```'} />
  }

  return (
    <React.Fragment>
      <Divider />
      <CardActions
        primaryActions={[
          {
            label: 'Generate Slack App Manifest',
            handleOnClick: () => {
              getManifest()
              setShowManifest(true)
            },
          },
        ]}
      />
      <Dialog
        classes={{
          paper: classes.dialog,
        }}
        open={showManifest}
        onClose={() => setShowManifest(false)}
        TransitionProps={{ onExited: () => setCopied(false) }}
        fullWidth
      >
        <DialogTitle>New Slack App Manifest</DialogTitle>
        <DialogContent>{renderContent()}</DialogContent>
        <DialogActions>
          <Button
            variant='contained'
            color='primary'
            onClick={() => {
              copyToClipboard(data.generateSlackAppManifest)
              setCopied(true)
            }}
            disabled={loading}
          >
            {copied ? 'Copied' : 'Copy'}
          </Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  )
}
