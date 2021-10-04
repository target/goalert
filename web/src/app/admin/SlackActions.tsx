import React, { useState } from 'react'
import makeStyles from '@mui/styles/makeStyles';
import Button from '@mui/material/Button'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import DialogContentText from '@mui/material/DialogContentText'
import DialogTitle from '@mui/material/DialogTitle'
import Divider from '@mui/material/Divider'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import { gql, useLazyQuery } from '@apollo/client'
import CardActions from '../details/CardActions'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import Markdown from '../util/Markdown'
import CopyText from '../util/CopyText'
import AppLink from '../util/AppLink'

const query = gql`
  query {
    generateSlackAppManifest
  }
`

const useStyles = makeStyles({
  copyButton: {
    float: 'right',
    padding: '12px',
  },
  dialog: {
    minHeight: '650px',
  },
})

export default function SlackActions(): JSX.Element {
  const classes = useStyles()
  const [showManifest, setShowManifest] = useState(false)

  const [getManifest, { called, loading, error, data }] = useLazyQuery(query, {
    pollInterval: 0,
  })

  function renderContent(): JSX.Element {
    if (called && loading) return <Spinner />
    if (error) return <GenericError error={error.message} />

    const manifest = data?.generateSlackAppManifest ?? ''
    return (
      <div>
        <div className={classes.copyButton}>
          <CopyText value={manifest} placement='left' />
        </div>
        <Markdown value={'```\n' + manifest + '\n```'} />
      </div>
    )
  }

  return (
    <React.Fragment>
      <Divider />
      <CardActions
        primaryActions={[
          {
            label: 'Create New Slack App',
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
        fullWidth
      >
        <DialogTitle data-cy='dialog-title'>Create New Slack App</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Copy the manifest generated below to configure a new GoAlert app
            within Slack.
          </DialogContentText>
          {renderContent()}
          <DialogContentText>
            Learn more about manifests{' '}
            <AppLink to='https://api.slack.com/reference/manifests' newTab>
              here
            </AppLink>
            .
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button color='primary' onClick={() => setShowManifest(false)}>
            Cancel
          </Button>
          <Button
            variant='contained'
            color='primary'
            endIcon={<OpenInNewIcon />}
            component={AppLink}
            to='https://api.slack.com/apps'
            newTab
            data-cy='configure-in-slack'
          >
            Configure in Slack
          </Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  )
}
