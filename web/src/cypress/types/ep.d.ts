// eslint-disable-next-line @typescript-eslint/no-unused-vars
namespace Cypress {
  interface Chainable {
    createEP: typeof createEP
    deleteEP: typeof deleteEP
    createEPStep: typeof createEPStep
  }
}

interface EP {
  id: string
  name: string
  description: string
  repeat: number
  stepCount: number
}

interface EPOptions {
  name?: string
  description?: string
  repeat?: number
  stepCount?: number
}

interface EPStep {
  delayMinutes: number
}

interface EPStepOptions {
  epID?: string
  ep?: EPOptions
  delay?: number
  targets?: [Target]
}
