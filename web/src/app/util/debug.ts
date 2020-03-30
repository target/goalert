export function debug(args: any): void {
  if (!window.console) return // some browsers don't define console if the devtools are closed
  if (console.debug) console.debug(args)
  // not all browsers have `.debug` defined
  else console.log(args)
}
export function warn(...args: any[]): void {
  if (!window.console) return // some browsers don't define console if the devtools are closed
  if (console.warn) console.warn(...args)
  // not all browsers have `.warn` defined
  else console.log(...args)
}
