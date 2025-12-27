import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { previewApi, createPreviewWebSocket } from '../api/client'
import type { PreviewResult, PreviewProgress, Message } from '../api/types'

export const usePreviewStore = defineStore('preview', () => {
  const result = ref<PreviewResult | null>(null)
  const progress = ref<PreviewProgress | null>(null)
  const messages = ref<Message[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const wsConnected = ref(false)

  let ws: WebSocket | null = null

  const matchedMessages = computed(() => messages.value.filter(m => m.matched_rule))

  const unmatchedMessages = computed(() => messages.value.filter(m => !m.matched_rule))

  const matchStats = computed(() => {
    if (!result.value) return null
    return {
      total: result.value.total_messages,
      matched: result.value.matched_messages,
      percentage:
        result.value.total_messages > 0
          ? Math.round((result.value.matched_messages / result.value.total_messages) * 100)
          : 0,
    }
  })

  async function fetchPreview(accountId: number, folder = 'INBOX', limit = 100) {
    loading.value = true
    error.value = null
    messages.value = []
    try {
      result.value = await previewApi.preview(accountId, folder, limit)
      messages.value = result.value.messages
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message
    } finally {
      loading.value = false
    }
  }

  async function applyRules(accountId: number, folder = 'INBOX', dryRun = false) {
    loading.value = true
    error.value = null
    try {
      result.value = await previewApi.apply(accountId, folder, dryRun)
      messages.value = result.value.messages
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message
    } finally {
      loading.value = false
    }
  }

  function startLivePreview(accountId: number, folder = 'INBOX', limit = 100) {
    // Close existing connection
    if (ws) {
      ws.close()
    }

    loading.value = true
    error.value = null
    messages.value = []
    progress.value = null
    result.value = null

    ws = createPreviewWebSocket()

    ws.onopen = () => {
      wsConnected.value = true
      ws?.send(
        JSON.stringify({
          type: 'preview',
          payload: {
            account_id: accountId,
            folder,
            limit,
          },
        })
      )
    }

    ws.onmessage = event => {
      try {
        const msg = JSON.parse(event.data)

        switch (msg.type) {
          case 'progress':
            progress.value = JSON.parse(msg.payload)
            if (progress.value?.message_data) {
              // Update or add message
              const msgData = progress.value.message_data
              const existingIndex = messages.value.findIndex(m => m.uid === msgData.uid)
              if (existingIndex >= 0) {
                messages.value[existingIndex] = msgData
              } else {
                messages.value.push(msgData)
              }
            }
            break

          case 'result':
            result.value = JSON.parse(msg.payload)
            messages.value = result.value?.messages || []
            loading.value = false
            break

          case 'error':
            error.value = msg.error
            loading.value = false
            break
        }
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e)
      }
    }

    ws.onerror = event => {
      console.error('WebSocket error:', event)
      error.value = 'WebSocket connection error'
      loading.value = false
    }

    ws.onclose = () => {
      wsConnected.value = false
    }
  }

  function stopLivePreview() {
    if (ws) {
      ws.close()
      ws = null
    }
    wsConnected.value = false
    loading.value = false
  }

  function clearPreview() {
    result.value = null
    messages.value = []
    progress.value = null
    error.value = null
  }

  return {
    result,
    progress,
    messages,
    loading,
    error,
    wsConnected,
    matchedMessages,
    unmatchedMessages,
    matchStats,
    fetchPreview,
    applyRules,
    startLivePreview,
    stopLivePreview,
    clearPreview,
  }
})
