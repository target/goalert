import React from 'react'
import { useTheme } from '@mui/material/styles'
import AutoSizer from 'react-virtualized-auto-sizer'
import { Bar, BarChart, CartesianGrid, Tooltip, XAxis, YAxis } from 'recharts'

interface ResponseBarGraphProps {
  data: { label: string; count: number }[]
  loading: boolean
}

export default function ResponseBarGraph(
  props: ResponseBarGraphProps,
): React.JSX.Element {
  const theme = useTheme()

  return (
    <div style={{ height: 300 }} data-cy='response-ack-graph'>
      <AutoSizer>
        {({ width, height }: { width: number; height: number }) => (
          <BarChart
            width={width}
            height={height}
            data={props.data}
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
              interval={0}
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
              formatter={(value: number) => [value, 'Alerts']}
            />
            <Bar
              dataKey='count'
              radius={[4, 4, 0, 0]}
              fill={
                props.loading
                  ? theme.palette.secondary.light
                  : theme.palette.primary.main
              }
            />
          </BarChart>
        )}
      </AutoSizer>
    </div>
  )
}
