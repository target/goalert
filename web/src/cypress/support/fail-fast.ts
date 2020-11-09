// NOTE the runner method is not inherent to Cypress
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const CY: any = Cypress

// Fail-fast-single-file
afterEach(function () {
  if (
    this.currentTest?.state === 'failed' &&
    // eslint-disable-next-line @typescript-eslint/ban-ts-ignore
    // @ts-ignore _currentRetry is private but required to work properly
    this.currentTest._currentRetry === this.currentTest.retries()
  ) {
    CY.runner.stop()
  }
})

export {}
