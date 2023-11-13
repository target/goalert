import React, { ReactNode } from 'react'
import { CardHeader, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'

const useStyles = makeStyles(() => ({
  headerNote: {
    fontStyle: 'italic',
  },
}))

export interface ListHeaderProps {
  // cardHeader will be displayed at the top of the card
  cardHeader?: ReactNode
  // header elements will be displayed at the top of the list.
  headerNote?: string // left-aligned
  headerAction?: React.ReactNode // right-aligned
}

export function ListHeader(props: ListHeaderProps): React.ReactNode {
  const classes = useStyles()
  const { headerNote, headerAction, cardHeader } = props
  return (
    <React.Fragment>
      {cardHeader}
      {(headerNote || headerAction) && (
        <CardHeader
          subheader={
            <Typography color='textSecondary' className={classes.headerNote}>
              {headerNote}
            </Typography>
          }
          action={headerAction}
        />
      )}
    </React.Fragment>
  )
}
