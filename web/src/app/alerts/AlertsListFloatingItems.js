import React, { useState } from 'react'
import SnackbarContent from '@material-ui/core/SnackbarContent'
import InfoIcon from '@material-ui/icons/Info'
import Snackbar from '@material-ui/core/Snackbar'
import { useQuery } from '@apollo/react-hooks'
import gql from 'graphql-tag'
import { isWidthDown, makeStyles } from '@material-ui/core'
import { useSelector } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { PropTypes as p } from 'prop-types'
import CreateAlertFab from './CreateAlertFab'
import useWidth from '../util/useWidth'

const useStyles = makeStyles(theme => ({
  snackbar: {
    backgroundColor: theme.palette.primary['500'],
    height: '6.75em',
    width: '20em', // only triggers on desktop, 100% on mobile devices
  },
  snackbarIcon: {
    fontSize: 20,
    opacity: 0.9,
    marginRight: theme.spacing(1),
  },
  snackbarMessage: {
    display: 'flex',
    alignItems: 'center',
  },
}))

/**
 * Handles rendering the floating page fab and snackbar warning
 * for the alerts list
 *
 * The page fab and snackbar are rendered together for proper
 * transitions on breakpoints md and down.
 */
export default function AlertsListFloatingItems(props) {
  const classes = useStyles()
  const width = useWidth()
  const isFullScreen = isWidthDown('md', width)

  // get redux vars
  const params = useSelector(urlParamSelector)
  const isFirstLogin = params('isFirstLogin')
  const allServices = params('allServices')

  // always open unless clicked away from or there are services present
  const [snackbarOpen, setSnackbarOpen] = useState(true)

  function handleCloseSnackbar(event, reason) {
    if (reason === 'clickaway') {
      setSnackbarOpen(false)
    }
  }

  // query to see if the current user has any favorited services
  // if allServices is not true
  const { loading, error, data } = useQuery(
    gql`
      query($input: ServiceSearchOptions) {
        services(input: $input) {
          nodes {
            id
          }
        }
      }
    `,
    {
      variables: {
        favoritesOnly: true,
        first: 1,
      },
    },
  )

  if (!data && (error || loading)) return null

  const noFavorites = !data?.nodes?.length
  const showFavoritesWarning =
    snackbarOpen &&
    !allServices &&
    !props.serviceID &&
    !isFirstLogin &&
    noFavorites

  return (
    <React.Fragment>
      <Snackbar
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
        open={showFavoritesWarning}
        onClose={handleCloseSnackbar}
      >
        <SnackbarContent
          className={classes.snackbar}
          aria-describedby='client-snackbar'
          message={
            <span id='client-snackbar' className={classes.snackbarMessage}>
              <InfoIcon className={classes.snackbarIcon} />
              It looks like you have no favorited services. Visit your most used
              services to set them as a favorite, or enable the filter to view
              alerts for all services.
            </span>
          }
        />
      </Snackbar>
      <CreateAlertFab
        serviceID={props.serviceID}
        showFavoritesWarning={showFavoritesWarning}
        transition={isFullScreen && (showFavoritesWarning || props.updateComplete)}
      />
    </React.Fragment>
  )
}

AlertsListFloatingItems.propTypes = {
  serviceID: p.string,
  updateComplete: p.bool,
}
