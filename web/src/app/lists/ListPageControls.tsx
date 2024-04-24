import React from 'react'
import {
  Button,
  Card,
  CircularProgress,
  Grid,
  IconButton,
  Theme,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Add, ChevronLeft, ChevronRight } from '@mui/icons-material'
import CreateFAB from './CreateFAB'
import { useIsWidthDown } from '../util/useWidth'

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
type ListPageControlsBaseProps = {
  onBack?: () => void
  onNext?: () => void
  loading?: boolean

  slots: {
    list: React.ReactNode
    search?: React.ReactNode
  }

  // ignored unless onCreateClick is provided
  createLabel?: string
}

type ListPageControlsCreatableProps = ListPageControlsBaseProps & {
  createLabel: string
  onCreateClick: () => void
}

export type ListPageControlsProps =
  | ListPageControlsBaseProps
  | ListPageControlsCreatableProps

function canCreate(
  props: ListPageControlsProps,
): props is ListPageControlsCreatableProps {
  return 'createLabel' in props
}

export default function ListPageControls(
  props: ListPageControlsProps,
): React.ReactNode {
  const classes = useStyles()
  const showCreate = canCreate(props)
  const isMobile = useIsWidthDown('md')

  return (
    <Grid container spacing={2}>
      <Grid
        container
        item
        xs={12}
        spacing={2}
        justifyContent='flex-start'
        alignItems='center'
      >
        {props.slots.search && <Grid item>{props.slots.search}</Grid>}

        {showCreate && !isMobile && (
          <Grid item sx={{ ml: 'auto' }}>
            <Button
              variant='contained'
              startIcon={<Add />}
              onClick={props.onCreateClick}
            >
              Create {props.createLabel}
            </Button>
          </Grid>
        )}
      </Grid>

      <Grid item xs={12}>
        <Card>{props.slots.list}</Card>
      </Grid>

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
            disabled={!props.onBack}
            onClick={() => {
              if (props.onBack) props.onBack()
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
            disabled={!props.onNext || props.loading}
            onClick={() => {
              if (props.onNext) props.onNext()
              window.scrollTo(0, 0)
            }}
          >
            {props.loading && (
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
      {showCreate && isMobile && (
        <React.Fragment>
          <CreateFAB
            onClick={props.onCreateClick}
            title={`Create ${props.createLabel}`}
          />
        </React.Fragment>
      )}
    </Grid>
  )
}
