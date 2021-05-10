import React, { cloneElement, ReactNode } from 'react'
import Card from '@material-ui/core/Card'
import CardHeader from '@material-ui/core/CardHeader'
import CardContent from '@material-ui/core/CardContent'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core/styles'
import CardActions, { Action } from '../details/CardActions'
import Markdown from '../util/Markdown'

export interface DataCardProps {
  title: string
  avatar?: JSX.Element // placement for an icon or image
  subheader?: string | JSX.Element
  details?: string
  primaryActions?: Array<Action | JSX.Element>
  secondaryActions?: Array<Action | JSX.Element>
}

const useStyles = makeStyles({
  flexHeight: {
    flexGrow: 1,
  },
  headerContent: {
    paddingTop: 0,
  },
})

export default function DataCard(p: DataCardProps): JSX.Element {
  const classes = useStyles()

  const avatar = (): ReactNode => {
    if (!p.avatar) return null
    return cloneElement(p.avatar, {
      style: { width: 56, height: 56 },
    })
  }

  return (
    <Card>
      <Grid item xs container direction='column'>
        <Grid item>
          <CardHeader
            title={p.title}
            subheader={p.subheader}
            avatar={avatar()}
            titleTypographyProps={{
              'data-cy': 'title',
              variant: 'h5',
              component: 'h2',
            }}
            subheaderTypographyProps={{
              'data-cy': 'subheader',
              variant: 'body1',
            }}
          />
        </Grid>

        {p.details && (
          <Grid item>
            <CardContent className={classes.headerContent}>
              <Typography
                component='div'
                variant='subtitle1'
                color='textSecondary'
                data-cy='details'
              >
                <Markdown value={p.details} />
              </Typography>
            </CardContent>
          </Grid>
        )}

        <Grid className={classes.flexHeight} item />
        {(p.primaryActions?.length || p.secondaryActions?.length) && (
          <Grid item>
            <CardActions
              primaryActions={p.primaryActions}
              secondaryActions={p.secondaryActions}
            />
          </Grid>
        )}
      </Grid>
    </Card>
  )
}
