import React from 'react'
import Grid, { GridProps } from '@material-ui/core/Grid'
import DataCard, { DataCardProps } from './DataCard'

interface CardCollectionProps {
  items: Array<DataCardProps>
  GridContainerProps?: GridProps
}

export default function CardCollection(p: CardCollectionProps): JSX.Element {
  return (
    <Grid container spacing={2} {...p.GridContainerProps}>
      {p.items.map((item, idx) => (
        <Grid key={item.title + '-' + idx} item>
          <DataCard {...item} />
        </Grid>
      ))}
    </Grid>
  )
}
