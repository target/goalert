import React, { useState } from 'react'
import Divider from '@material-ui/core/Divider'
import CardActions from '../details/CardActions'
import Dialog from '@material-ui/core/Dialog'
import { gql, useQuery } from '@apollo/client'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import Markdown from '../util/Markdown'
import {
  Button,
  DialogActions,
  DialogContent,
  DialogTitle,
} from '@material-ui/core'
import copyToClipboard from '../util/copyToClipboard'

const query = gql`
  query {
    generateSlackAppManifest
  }
`

export default function SlackActions(): JSX.Element {
  const [showManifest, setShowManifest] = useState(false)
  const [copied, setCopied] = useState(false)

  const { loading, error, data } = useQuery(query)
  if (loading) return <Spinner />
  if (error) return <GenericError error={error.message} />

  return (
    <React.Fragment>
      <Divider />
      <CardActions
        primaryActions={[
          {
            label: 'Generate Slack App Manifest',
            handleOnClick: () => setShowManifest(true),
          },
        ]}
      />
      <Dialog open={showManifest} onClose={() => setShowManifest(false)}>
        <DialogTitle>New Slack App Manifest</DialogTitle>
        <DialogContent>
          <Markdown value={data.generateSlackAppManifest} />
        </DialogContent>
        <DialogActions>
          <Button
            variant='contained'
            color='primary'
            onClick={() => {
              copyToClipboard(data.generateSlackAppManifest)
              setCopied(true)
            }}
          >
            {copied ? 'Copied' : 'Copy'}
          </Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  )
}
