import React, { useEffect, useState } from 'react'
import Grid from '@mui/material/Grid'
import { gql, useMutation, useQuery } from 'urql'
import { DateTime } from 'luxon'
import { SWOAction, SWONode as SWONodeType, SWOStatus } from '../../../schema'
import Notices, { Notice } from '../../details/Notices'
import SWONode from './SWONode'
import Spinner from '../../loading/components/Spinner'
import AdminSWOConfirmDialog from './AdminSWOConfirmDialog'
import { errCheck, friendlyName, toTitle } from './util'
import { AdminSWODone } from './AdminSWODone'
import { AdminSWOWrongMode } from './AdminSWOWrongMode'
import { AdminSWODBVersionCard } from './AdminSWODBVersionCard'
import { AdminSWOStatusCard } from './AdminSWOStatusCard'
import { AdminSwitchoverGuideButton } from './AdminSwitchoverGuide'

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

const mutation = gql`
  mutation ($action: SWOAction!) {
    swoAction(action: $action)
  }
`

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
  if (data?.state === 'done')
    return (
      <React.Fragment>
        <AdminSWODone />
        <center style={{ paddingTop: '1em' }}>
          <AdminSwitchoverGuideButton />
        </center>
      </React.Fragment>
    )

  if (error && error.message === '[GraphQL] not in SWO mode' && !data)
    return (
      <React.Fragment>
        <AdminSWOWrongMode />
        <center style={{ paddingTop: '1em' }}>
          <AdminSwitchoverGuideButton />
        </center>
      </React.Fragment>
    )

  if (!data)
    return (
      <React.Fragment>
        <Spinner />
        <center style={{ paddingTop: '1em' }}>
          <AdminSwitchoverGuideButton />
        </center>
      </React.Fragment>
    )

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
      details: toTitle(mutationStatus.error.message),
      endNote: DateTime.local().toFormat('fff'),
    })
  }
  if (error && error.message !== '[GraphQL] not in SWO mode') {
    statusNotices.push({
      type: 'error',
      message: 'Failed to fetch status',
      details: toTitle(error.message),
      endNote: DateTime.local().toFormat('fff'),
    })
  }

  const configErr = errCheck(data)

  return (
    <Grid container spacing={2}>
      {showConfirm && (
        <AdminSWOConfirmDialog
          messages={configErr}
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
          onExecClick={() => {
            if (configErr.length) {
              setShowConfirm(true)
              return false
            }

            actionHandler('execute')()
            return true
          }}
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
