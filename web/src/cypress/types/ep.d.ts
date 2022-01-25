declare namespace Cypress {
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
  isFavorite: boolean
}

interface EPOptions {
  name?: string
  description?: string
  repeat?: number
  stepCount?: number
  favorite?: boolean
}

interface EPStep {
  id: string
  stepNumber: number
  delayMinutes: number
  targets: Target[]
  escalationPolicy: EP
}

interface EPStepOptions {
  epID?: string
  ep?: EPOptions
  delay?: number
  targets?: [Target]
}
