import React from 'react'
import { CircularProgress, Grid, IconButton, Theme } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { ChevronLeft, ChevronRight } from '@mui/icons-material'
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
  loadMore?: (numberToLoad?: number) => void
  pageCount: number
  page: number
  setPage: (page: number) => void
  isLoading: boolean
}): JSX.Element {
  const classes = useStyles()
  const { loadMore, pageCount, page, setPage, isLoading } = props

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
            if (onBack) onBack()
            window.scrollTo(0, 0)
          }}
        >
          <ChevronLeft />
        </IconButton>
      </Grid>
      <Grid item>
        <IconButton
          title='next page'
          data-cy='next-button'
          disabled={!onNext}
          onClick={() => {
            if (onNext) onNext()
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
          <ChevronRight />
        </IconButton>
      </Grid>
    </Grid>
  )
}
