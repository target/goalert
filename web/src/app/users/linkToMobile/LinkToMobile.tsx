import React, { useEffect, useState, ReactNode } from 'react'
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  makeStyles,
  isWidthDown,
  DialogContentText,
} from '@material-ui/core'
import PhonelinkIcon from '@material-ui/icons/Phonelink'
import SwipeableViews from 'react-swipeable-views'
import { virtualize, bindKeyboard } from 'react-swipeable-views-utils'
import gql from 'graphql-tag'
import { useQuery, useMutation } from 'react-apollo'
import DialogTitleWrapper from '../../dialogs/components/DialogTitleWrapper'
import useWidth from '../../util/useWidth'
import { styles as globalStyles } from '../../styles/materialStyles'
import ClaimCodeDisplay from './ClaimCodeDisplay'
import VerifyCodeFields from './VerifyCodeFields'
import Spinner from '../../loading/components/Spinner'
import SuccessAnimation from '../../util/animations/SuccessAnimation'
import ErrorAnimation from '../../util/animations/ErrorAnimation'
import { useDispatch, useSelector } from 'react-redux'
import { urlParamSelector } from '../../selectors'
import { setURLParam } from '../../actions'
import DialogContentError from '../../dialogs/components/DialogContentError'

interface SlideParams {
  index: number
  key: number
}

const VirtualizeAnimatedViews = bindKeyboard(virtualize(SwipeableViews))

const useStyles = makeStyles((theme) => {
  const { cancelButton } = globalStyles(theme)

  return {
    cancelButton,
    dialog: {
      height: 345,
      justifyContent: 'space-between',
    },
    successContainer: {
      height: '100%',
      width: '100%',
      display: 'flex',
      alignItems: 'center',
      textAlign: 'center',
    },
  }
})

const mutation = gql`
  mutation {
    createAuthLink {
      id
      claimCode
      verifyCode
    }
  }
`

export const query = gql`
  query authLinkStatus($id: ID!) {
    authLinkStatus(id: $id) {
      id
      expiresAt
      claimed
      verified
      authed
    }
  }
`

export default function LinkToMobile(): JSX.Element {
  const classes = useStyles()
  const width = useWidth()
  const fullscreen = isWidthDown('md', width)
  const [showDialog, setShowDialog] = useState(false)

  // getter for error messages
  const urlParams = useSelector(urlParamSelector)
  const errorMsg = urlParams('error', '')

  // setter for error messages
  const dispatch = useDispatch()
  const setErrorMessage = (value: string): void => {
    dispatch(setURLParam('error', value))
  }

  const [createAuthLink, createAuthLinkStatus] = useMutation(mutation, {
    onError: (err) => {
      if (err.message) setErrorMessage(err.message)
    },
  })
  const loading = !createAuthLinkStatus.data && createAuthLinkStatus.loading
  const authLinkID = createAuthLinkStatus?.data?.createAuthLink.id ?? ''
  const claimCode = createAuthLinkStatus?.data?.createAuthLink.claimCode ?? ''
  const verifyCode = createAuthLinkStatus?.data?.createAuthLink.verifyCode ?? ''

  const { data, error } = useQuery(query, {
    variables: {
      id: authLinkID,
    },
    skip: loading || !authLinkID,
  })

  // set error if query fails
  const queryErrorMsg = error?.message ?? ''
  useEffect(() => {
    if (queryErrorMsg) setErrorMessage(queryErrorMsg)
  }, [queryErrorMsg])

  const claimed = data?.authLinkStatus?.claimed ?? ''
  const authed = data?.authLinkStatus?.authed ?? ''

  useEffect(() => {
    if (showDialog) {
      createAuthLink()
    }
  }, [showDialog])

  // index of stepper/slide
  let index = 0
  const LAST_STEP_IDX = 2
  if (claimed && index !== 1) index = 1
  if (authed && index !== 2) index = 2

  function slideRenderer({ index, key }: SlideParams): ReactNode {
    switch (index) {
      case 0:
        return (
          <ClaimCodeDisplay
            key={key}
            authLinkID={authLinkID}
            claimCode={claimCode}
          />
        )
      case 1:
        return (
          <VerifyCodeFields
            key={key}
            authLinkID={authLinkID}
            verifyCode={verifyCode}
          />
        )
      case 2:
        return <Success key={key} isStopped={!authed} />
      default:
        return null
    }
  }

  // Render the main link to mobile content
  function renderDialogContent(): ReactNode {
    if (errorMsg) return null

    return (
      <React.Fragment>
        <DialogTitleWrapper title='Link to Mobile' />
        {loading ? (
          <DialogContent>
            <Spinner />
          </DialogContent>
        ) : (
          <VirtualizeAnimatedViews
            index={index}
            onChangeIndex={(i: number) => (index = i)}
            slideRenderer={slideRenderer}
          />
        )}
        <DialogActions>
          {index === LAST_STEP_IDX ? (
            <Button
              variant='contained'
              color='primary'
              onClick={() => setShowDialog(false)}
            >
              Done
            </Button>
          ) : (
            <Button
              className={classes.cancelButton}
              onClick={() => setShowDialog(false)}
            >
              Cancel
            </Button>
          )}
        </DialogActions>
      </React.Fragment>
    )
  }

  // render an error message if seen in the url parameter
  function renderDialogError(): ReactNode {
    if (!errorMsg) return null

    return (
      <React.Fragment>
        <DialogTitleWrapper title='An error occurred' />
        <DialogContent>
          <DialogContentText>
            Whoops, looks like something went wrong. Please try again.
          </DialogContentText>
        </DialogContent>
        <ErrorAnimation autoplay />
        <DialogContentError error={errorMsg} noPadding />
        <DialogActions>
          <Button
            color='primary'
            variant='contained'
            onClick={() => setShowDialog(false)}
          >
            Okay
          </Button>
        </DialogActions>
      </React.Fragment>
    )
  }

  // button on profile and dialog parent
  return (
    <React.Fragment>
      <Button
        color='primary'
        variant='contained'
        startIcon={<PhonelinkIcon />}
        onClick={() => setShowDialog(true)}
      >
        Link to Mobile
      </Button>

      <Dialog
        classes={{ paper: classes.dialog }}
        open={showDialog}
        fullScreen={fullscreen}
        fullWidth
        maxWidth='xs'
        onClose={() => setShowDialog(false)}
        onExited={() => setErrorMessage('')}
      >
        {renderDialogContent()}
        {renderDialogError()}
      </Dialog>
    </React.Fragment>
  )
}

interface SuccessProps {
  isStopped: boolean
}

// renders a centered success animation
function Success(props: SuccessProps): JSX.Element {
  const classes = useStyles()
  return (
    <div className={classes.successContainer}>
      <DialogContent>
        <SuccessAnimation isStopped={props.isStopped} />
        <DialogContentText>Success!</DialogContentText>
      </DialogContent>
    </div>
  )
}
