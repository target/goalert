import React from 'react'
import { Grid } from '@mui/material'
import { useTheme } from '@mui/material/styles'
import AutoSizer from 'react-virtualized-auto-sizer'
import { Bar, BarChart, CartesianGrid, Tooltip, XAxis, YAxis } from 'recharts'

interface ResponseBarGraphProps {
  data: { label: string; count: number }[]
  loading: boolean
  height?: number
  // Show every x-axis label (for small categorical data) instead of thinning
  // them out to avoid overlap.
  showAllLabels?: boolean
  'data-cy'?: string
}

export default function ResponseBarGraph(
  props: ResponseBarGraphProps,
): React.JSX.Element {
  const theme = useTheme()
  const { data, loading, height = 300 } = props

  return (
    <Grid
      container
      sx={{
        height: `${height}px`,
        fontFamily: theme.typography.body2.fontFamily,
      }}
    >
      <Grid item xs={12} data-cy={props['data-cy']}>
        <AutoSizer>
          {({ width, height }: { width: number; height: number }) => (
            <BarChart
              width={width}
              height={height}
              data={data}
              margin={{ top: 16, right: 16, bottom: 8, left: -16 }}
            >
              <CartesianGrid
                strokeDasharray='3 3'
                vertical={false}
                stroke={theme.palette.divider}
              />
              <XAxis
                dataKey='label'
                stroke={theme.palette.text.secondary}
                interval={props.showAllLabels ? 0 : 'preserveStartEnd'}
                tickLine={false}
                fontSize={12}
              />
              <YAxis
                allowDecimals={false}
                stroke={theme.palette.text.secondary}
                tickLine={false}
              />
              <Tooltip
                cursor={{ fill: theme.palette.action.hover }}
                contentStyle={{
                  background: theme.palette.background.paper,
                  border: `1px solid ${theme.palette.divider}`,
                  borderRadius: 8,
                }}
                labelStyle={{ color: theme.palette.text.primary }}
                itemStyle={{ color: theme.palette.text.secondary }}
                formatter={(value: number) => [value, 'Alerts']}
              />
              <Bar
                dataKey='count'
                name='Alerts'
                radius={[4, 4, 0, 0]}
                fill={
                  loading
                    ? theme.palette.secondary.light
                    : theme.palette.primary.main
                }
              />
            </BarChart>
          )}
        </AutoSizer>
      </Grid>
    </Grid>
  )
}
