import React, { useEffect, useState } from 'react'
import Grid from '@mui/material/Grid'
import { gql, useMutation, useQuery } from 'urql'
import { DateTime } from 'luxon'
import { SWOAction, SWONode as SWONodeType, SWOStatus } from '../../../schema'
import Notices, { Notice } from '../../details/Notices'
import SWONode from './SWONode'
import Spinner from '../../loading/components/Spinner'
import AdminSWOConfirmDialog from './AdminSWOConfirmDialog'
import { errCheck } from './errCheck'
import { AdminSWODone } from './AdminSWODone'
import { AdminSWOWrongMode } from './AdminSWOWrongMode'
import { AdminSWODBVersionCard } from './AdminSWODBVersionCard'
import { AdminSWOStatusCard } from './AdminSWOStatusCard'

const query = gql`
  query {
    swoStatus {
      state
      lastError
      lastStatus
      mainDBVersion
      nextDBVersion
      nodes {
        id
        uptime
        canExec
        isLeader
        configError
        connections {
          name
          version
          type
          isNext
          count
        }
      }
    }
  }
`

let n = 1
let u = 1
const names: { [key: string]: string } = {}

// friendlyName will assign a persistant "friendly" name to the node.
//
// This ensures a specific ID will always refer to the same node. This
// is so that it is clear if a node dissapears or a new one appears.
//
// Note: `Node 1` on one browser tab may not be the same node as `Node 1`
// on another browser tab.
function friendlyName(id: string): string {
  if (!names[id]) {
    if (id.startsWith('unknown')) return (names[id] = 'Unknown ' + u++)
    return (names[id] = 'Node ' + n++)
  }
  return names[id]
}

const mutation = gql`
  mutation ($action: SWOAction!) {
    swoAction(action: $action)
  }
`

function cptlz(s: string): string {
  return s.charAt(0).toUpperCase() + s.substring(1)
}

export default function AdminSwitchover(): JSX.Element {
  const [{ fetching, error, data: _data }, refetch] = useQuery({
    query,
  })

  const [showConfirm, setShowConfirm] = useState(false)

  const [mutationStatus, commit] = useMutation(mutation)
  const data = _data?.swoStatus as SWOStatus

  useEffect(() => {
    if (data?.state === 'done') return
    if (mutationStatus.fetching) return

    const t = setInterval(() => {
      if (!fetching) refetch()
    }, 1000)
    return () => clearInterval(t)
  }, [fetching, refetch, data?.state, mutationStatus.fetching])

  // remember if we are done and stay that way
  if (data?.state === 'done') return <AdminSWODone />

  if (error && error.message === '[GraphQL] not in SWO mode' && !data)
    return <AdminSWOWrongMode />

  if (!data) return <Spinner />

  function actionHandler(action: 'reset' | 'execute'): () => void {
    return () => {
      commit({ action }, { additionalTypenames: ['SWOStatus'] })
    }
  }
  const statusNotices: Notice[] = []
  if (mutationStatus.error) {
    const vars: { action?: SWOAction } = mutationStatus.operation
      ?.variables || {
      action: '',
    }
    statusNotices.push({
      type: 'error',
      message: 'Failed to ' + vars.action,
      details: cptlz(mutationStatus.error.message),
      endNote: DateTime.local().toFormat('fff'),
    })
  }
  if (error && error.message !== '[GraphQL] not in SWO mode') {
    statusNotices.push({
      type: 'error',
      message: 'Failed to fetch status',
      details: cptlz(error.message),
      endNote: DateTime.local().toFormat('fff'),
    })
  }

  const configErr = errCheck(data).join('\n')

  return (
    <Grid container spacing={2}>
      {showConfirm && (
        <AdminSWOConfirmDialog
          message={configErr}
          onClose={() => setShowConfirm(false)}
          onConfirm={actionHandler('execute')}
        />
      )}
      {statusNotices.length > 0 && (
        <Grid item xs={12}>
          <Notices notices={statusNotices.reverse()} />
        </Grid>
      )}
      <Grid item xs={12} sm={12} md={12} lg={4} xl={4}>
        <AdminSWOStatusCard
          data={data}
          onExecClick={actionHandler('execute')}
          onResetClick={actionHandler('reset')}
        />
      </Grid>

      <Grid item xs={12} sm={12} md={12} lg={8} xl={8}>
        <AdminSWODBVersionCard data={data} />
      </Grid>

      <Grid item xs={12} container spacing={2} justifyContent='space-between'>
        {data?.nodes.length > 0 &&
          data.nodes
            .slice()
            .sort((a: SWONodeType, b: SWONodeType) => {
              const aName = friendlyName(a.id)
              const bName = friendlyName(b.id)
              if (aName < bName) return -1
              if (aName > bName) return 1
              return 0
            })
            .map((node: SWONodeType) => (
              <SWONode key={node.id} node={node} name={friendlyName(node.id)} />
            ))}
      </Grid>
    </Grid>
  )
}
