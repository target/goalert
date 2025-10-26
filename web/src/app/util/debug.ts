/* eslint-disable @typescript-eslint/no-explicit-any */

export function debug(message?: any, ...optionalParams: any[]): void {
  if (!window.console) return // some browsers don't define console if the devtools are closed
  if (console.debug) console.debug(message, ...optionalParams)
  // not all browsers have `.debug` defined
  else console.log(message, ...optionalParams)
}

export function warn(message?: any, ...optionalParams: any[]): void {
  if (!window.console) return // some browsers don't define console if the devtools are closed
  if (console.warn) console.warn(message, ...optionalParams)
  // not all browsers have `.warn` defined
  else console.log(message, ...optionalParams)
}
