// eslint-disable-next-line @typescript-eslint/no-unused-vars
namespace Cypress {
  interface Chainable {
    /** Click the FAB (floating action button) of the page.
     *
     * If the FAB is a Speed-Dial variant, you can optionally
     * specify the option label to select as an argument.
     */
    pageFab: typeof pageFab
  }
}
