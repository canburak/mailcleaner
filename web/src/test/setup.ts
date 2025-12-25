import { config } from '@vue/test-utils'

// Global test setup for Vue Test Utils

// Mock global components if needed
config.global.stubs = {
  // Add component stubs here if needed
}

// Mock window.matchMedia for responsive tests
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  }),
})
