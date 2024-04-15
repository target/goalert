import { execSync } from 'node:child_process'
import { promisify } from 'util'

/**
 * Returns array of PIDs that were started by the given parent PID.
 */
export const getChildProcesses = (parentPid: number): number[] => {
  try {
    // Get all running processes
    const allProcesses = execSync('ps -e -o pid= -o ppid=')
      .toString()
      .trim()
      .split('\n')

    // Filter out processes whose PPID matches the given PID
    const childPids = allProcesses
      .map((line) => line.trim().split(/\s+/))
      .filter(([, ppid]) => Number(ppid) === parentPid)
      .map(([pid]) => Number(pid))

    return childPids
  } catch (e) {
    console.error('Failed to get child processes', e)
    return []
  }
}

/**
 * Returns the command line binary and args.
 */
export const getProcessBinary = (pid: number): string => {
  try {
    const binary = execSync(`ps -p ${pid} -o command=`).toString().trim()
    return binary
  } catch (e) {
    return 'unknown'
  }
}

const sleep = promisify(setTimeout)

/**
 * Tries to kill the given PID with the given signal.
 * Returns true if the signal was sent successfully.
 */
const tryKill = (pid: number, signal: NodeJS.Signals): boolean => {
  try {
    process.kill(pid, signal)
    return true
  } catch (e) {
    return false
  }
}

export default async function globalTeardown(): Promise<void> {
  const shellPIDs = getChildProcesses(process.pid).filter((pid) =>
    getProcessBinary(pid).includes('./bin/goalert.cover'),
  )

  const goalertPIDs = shellPIDs.flatMap((pid) => getChildProcesses(pid))
  console.log('Found GoAlert PIDs:', goalertPIDs.map(getProcessBinary))

  // for loop over goalertPIDs, and send SIGTERM to each process.

  // eslint-disable-next-line no-labels
  nextPID: for (const pid of goalertPIDs) {
    console.log(`Killing process ${pid}...`)
    let limit = 5
    while (tryKill(pid, 'SIGINT')) {
      if (limit-- <= 0) {
        console.log(`Process ${pid} did not exit after 5 seconds, skipping...`)

        // eslint-disable-next-line no-labels
        continue nextPID
      }
      console.log(`Waiting for process ${pid} to exit...`)
      await sleep(100)
    }

    console.log(`Process ${pid} exited.`)
  }
}
