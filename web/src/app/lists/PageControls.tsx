import React from 'react'
import CircularProgress from '@material-ui/core/CircularProgress'
import Grid from '@material-ui/core/Grid'
import IconButton from '@material-ui/core/IconButton'
import { makeStyles, Theme } from '@material-ui/core'
import LeftIcon from '@material-ui/icons/ChevronLeft'
import RightIcon from '@material-ui/icons/ChevronRight'
import { ITEMS_PER_PAGE } from '../config'

const useStyles = makeStyles((theme: Theme) => ({
  progress: {
    color: theme.palette.secondary.main,
    position: 'absolute',
  },
  controls: {
    [theme.breakpoints.down('sm')]: {
      '&:not(:first-child)': {
        marginBottom: '4.5em',
        paddingBottom: '1em',
      },
    },
  },
}))

export function PageControls(props: {
  isLoading?: boolean
  loadMore?: (numberToLoad?: number) => void
  pageCount: number
  page: number
  setPage: (page: number) => void
}): JSX.Element {
  const classes = useStyles()
  const { isLoading, loadMore, pageCount, page, setPage } = props

  const hasNextPage = (() => {
    const nextPage = page + 1

    // Check that we have at least 1 item already for the next page
    if (nextPage < pageCount) return true

    // If we're on the last page, not already loading, and can load more
    if (nextPage === pageCount && !isLoading && loadMore) {
      return true
    }

    return false
  })()

  function handleNextPage(): void {
    const nextPage = page + 1
    setPage(nextPage)

    // If we're on a not-fully-loaded page, or the last page when > the first page
    if (
      (nextPage >= pageCount || (nextPage > 1 && nextPage + 1 === pageCount)) &&
      loadMore
    )
      loadMore(ITEMS_PER_PAGE * 2)
  }

  const onBack = page > 0 ? () => setPage(page - 1) : undefined
  const onNext = hasNextPage ? handleNextPage : undefined

  return (
    <Grid
      item // item within main render grid
      xs={12}
      container // container for control items
      spacing={1}
      justifyContent='flex-end'
      alignItems='center'
      className={classes.controls}
    >
      <Grid item>
        <IconButton
          title='back page'
          data-cy='back-button'
          disabled={!onBack}
          onClick={() => {
            onBack && onBack()
            window.scrollTo(0, 0)
          }}
        >
          <LeftIcon />
        </IconButton>
      </Grid>
      <Grid item>
        <IconButton
          title='next page'
          data-cy='next-button'
          disabled={!onNext}
          onClick={() => {
            onNext && onNext()
            window.scrollTo(0, 0)
          }}
        >
          {isLoading && !onNext && (
            <CircularProgress
              color='secondary'
              size={24}
              className={classes.progress}
            />
          )}
          <RightIcon />
        </IconButton>
      </Grid>
    </Grid>
  )
}
