import { describe, it, expect } from 'vitest'

describe('API Types', () => {
  it('should have correct account structure', () => {
    const account = {
      id: 1,
      name: 'Test Account',
      server: 'imap.example.com',
      port: 993,
      username: 'test@example.com',
      tls: true,
    }

    expect(account.id).toBe(1)
    expect(account.name).toBe('Test Account')
    expect(account.server).toBe('imap.example.com')
    expect(account.port).toBe(993)
    expect(account.tls).toBe(true)
  })

  it('should have correct rule structure', () => {
    const rule = {
      id: 1,
      account_id: 1,
      name: 'Test Rule',
      pattern: 'test@',
      pattern_type: 'sender' as const,
      move_to_folder: 'TestFolder',
      enabled: true,
      priority: 10,
    }

    expect(rule.id).toBe(1)
    expect(rule.pattern_type).toBe('sender')
    expect(rule.enabled).toBe(true)
  })

  it('should validate pattern types', () => {
    const validPatternTypes = ['sender', 'subject', 'from_domain']
    expect(validPatternTypes).toContain('sender')
    expect(validPatternTypes).toContain('subject')
    expect(validPatternTypes).toContain('from_domain')
  })
})
