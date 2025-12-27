import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { rulesApi } from '../api/client'
import type { Rule, RuleCreate } from '../api/types'

export const useRulesStore = defineStore('rules', () => {
  const rules = ref<Rule[]>([])
  const currentRule = ref<Rule | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const sortedRules = computed(() => [...rules.value].sort((a, b) => b.priority - a.priority))

  const enabledRules = computed(() => sortedRules.value.filter(r => r.enabled))

  async function fetchRules(accountId: number) {
    loading.value = true
    error.value = null
    try {
      rules.value = await rulesApi.list(accountId)
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message
    } finally {
      loading.value = false
    }
  }

  async function fetchRule(id: number) {
    loading.value = true
    error.value = null
    try {
      currentRule.value = await rulesApi.get(id)
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message
    } finally {
      loading.value = false
    }
  }

  async function createRule(accountId: number, data: RuleCreate): Promise<Rule | null> {
    loading.value = true
    error.value = null
    try {
      const rule = await rulesApi.create(accountId, data)
      rules.value.push(rule)
      return rule
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message
      return null
    } finally {
      loading.value = false
    }
  }

  async function updateRule(id: number, data: Partial<RuleCreate>): Promise<Rule | null> {
    loading.value = true
    error.value = null
    try {
      const rule = await rulesApi.update(id, data)
      const index = rules.value.findIndex(r => r.id === id)
      if (index >= 0) {
        rules.value[index] = rule
      }
      if (currentRule.value?.id === id) {
        currentRule.value = rule
      }
      return rule
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message
      return null
    } finally {
      loading.value = false
    }
  }

  async function deleteRule(id: number): Promise<boolean> {
    loading.value = true
    error.value = null
    try {
      await rulesApi.delete(id)
      rules.value = rules.value.filter(r => r.id !== id)
      if (currentRule.value?.id === id) {
        currentRule.value = null
      }
      return true
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message
      return false
    } finally {
      loading.value = false
    }
  }

  async function toggleRule(id: number): Promise<boolean> {
    const rule = rules.value.find(r => r.id === id)
    if (!rule) return false

    const updated = await updateRule(id, { enabled: !rule.enabled })
    return updated !== null
  }

  function clearRules() {
    rules.value = []
    currentRule.value = null
  }

  return {
    rules,
    currentRule,
    loading,
    error,
    sortedRules,
    enabledRules,
    fetchRules,
    fetchRule,
    createRule,
    updateRule,
    deleteRule,
    toggleRule,
    clearRules,
  }
})
