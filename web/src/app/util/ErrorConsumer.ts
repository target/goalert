import type { CombinedError } from 'urql'

type ErrorConsumerError = {
  path: string
  code: string
  fieldID: string
  message: string
}
type ErrorStore = {
  errors: Set<ErrorConsumerError>
}

export class ErrorConsumer {
  constructor(e?: CombinedError | null | undefined) {
    if (!e) return

    if (e.networkError) {
      this.store.errors.add({
        message: e.networkError.message,
        code: '',
        path: '',
        fieldID: '',
      })
    }

    e.graphQLErrors?.forEach((e) => {
      this.store.errors.add({
        message: e.message,
        code: e.extensions?.code?.toString() || '',
        path: e.path?.join('.') || '',
        fieldID: e.extensions?.fieldID?.toString() || '',
      })
    })

    this.hadErrors = this.hasErrors()

    // Use FinalizationRegistry if available, this will allow us to raise any errors that are forgotten.
    if (
      'FinalizationRegistry' in window &&
      typeof window.FinalizationRegistry === 'function'
    ) {
      //@ts-ignore
      const r = new window.FinalizationRegistry((e: { store: ErrorStore }) => {
        if (e.store.errors.size === 0) return
        e.store.errors.forEach((e) => console.error(e))
      })

      r.register(this, { e: this.store }, this)
    }
  }

  private isDone: boolean = false
  private store: ErrorStore = { errors: new Set() }

  /** Whether there were any errors in the original error. */
  public readonly hadErrors: boolean = false

  private doneCheck() {
    if (!this.isDone) return

    throw new Error(
      'ErrorConsumer is already done, ensure you are not calling this method after calling done() or remaining()',
    )
  }

  /** Returns and consumes (if exists) a single INVALID_INPUT_VALUE error with the given path. */
  getInputError(path: string): string | undefined {
    this.doneCheck()

    let result: string | undefined = undefined

    // find the first error with the given path
    this.store.errors.forEach((e) => {
      if (e.code !== 'INVALID_INPUT_VALUE') return
      if (e.path !== path) return
      if (result !== undefined) return

      result = e.message
      this.store.errors.delete(e)
    })

    return result
  }

  /** Returns and consumes (if exists) all INVALID_DEST_FIELD_VALUE errors.
   *
   * @param pathPrefix - If provided, only errors with the given path prefix will be consumed.
   */
  getAllDestFieldErrors(pathPrefix?: string): Readonly<Record<string, string>> {
    this.doneCheck()

    const errs: Record<string, string> = {}
    this.store.errors.forEach((e) => {
      if (e.code !== 'INVALID_DEST_FIELD_VALUE') return
      if (pathPrefix !== undefined && !e.path.startsWith(pathPrefix)) return
      if (errs[e.fieldID] !== undefined) return

      errs[e.fieldID] = e.message
      this.store.errors.delete(e)
    })

    return errs
  }

  /** Returns whether there are any errors remaining. */
  hasErrors(): boolean {
    return this.store.errors.size > 0
  }

  /** Returns and consumes all remaining errors. */
  remaining(): Readonly<Array<string>> {
    this.doneCheck()

    const errs: string[] = []
    this.store.errors.forEach((e) => {
      errs.push(e.message)
      this.store.errors.delete(e)
    })

    // mark as done
    this.isDone = true

    return errs
  }

  /** Logs and consumes any remaining errors. */
  done(): void {
    this.doneCheck()
    this.remaining().forEach((e) => console.error(e))
  }
}
