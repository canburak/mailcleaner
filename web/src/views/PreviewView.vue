<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAccountsStore } from '../stores/accounts'
import { useRulesStore } from '../stores/rules'
import { usePreviewStore } from '../stores/preview'

const props = defineProps<{ id: string }>()
const router = useRouter()
const accountsStore = useAccountsStore()
const rulesStore = useRulesStore()
const previewStore = usePreviewStore()

const accountId = ref(parseInt(props.id))
const selectedFolder = ref('INBOX')
const messageLimit = ref(100)
const useLivePreview = ref(true)
const filterMatched = ref<'all' | 'matched' | 'unmatched'>('all')

const displayedMessages = computed(() => {
  if (filterMatched.value === 'matched') {
    return previewStore.matchedMessages
  } else if (filterMatched.value === 'unmatched') {
    return previewStore.unmatchedMessages
  }
  return previewStore.messages
})

onMounted(async () => {
  await accountsStore.fetchAccount(accountId.value)
  await accountsStore.fetchFolders(accountId.value)
  await rulesStore.fetchRules(accountId.value)
})

onUnmounted(() => {
  previewStore.stopLivePreview()
})

watch(
  () => props.id,
  async newId => {
    accountId.value = parseInt(newId)
    previewStore.clearPreview()
    await accountsStore.fetchAccount(accountId.value)
    await accountsStore.fetchFolders(accountId.value)
    await rulesStore.fetchRules(accountId.value)
  }
)

function goBack() {
  router.push(`/accounts/${accountId.value}`)
}

async function runPreview() {
  if (useLivePreview.value) {
    previewStore.startLivePreview(accountId.value, selectedFolder.value, messageLimit.value)
  } else {
    await previewStore.fetchPreview(accountId.value, selectedFolder.value, messageLimit.value)
  }
}

function stopPreview() {
  previewStore.stopLivePreview()
}

async function applyRules() {
  if (confirm('This will move matching emails to their target folders. Continue?')) {
    await previewStore.applyRules(accountId.value, selectedFolder.value, false)
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString()
}
</script>

<template>
  <div>
    <div class="page-header">
      <div class="flex items-center gap-4">
        <button class="btn btn-outline" @click="goBack">&larr; Back</button>
        <h1 class="page-title">Live Preview</h1>
      </div>
    </div>

    <div v-if="previewStore.error" class="alert alert-error">
      {{ previewStore.error }}
    </div>

    <div class="card mb-4">
      <h3 class="card-title">Preview Settings</h3>
      <div class="preview-controls">
        <div class="form-group">
          <label class="form-label">Folder</label>
          <select v-model="selectedFolder" class="form-select">
            <option v-for="folder in accountsStore.folders" :key="folder.name" :value="folder.name">
              {{ folder.name }}
            </option>
            <option v-if="accountsStore.folders.length === 0" value="INBOX">INBOX</option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label">Message Limit</label>
          <select v-model.number="messageLimit" class="form-select">
            <option :value="25">25 messages</option>
            <option :value="50">50 messages</option>
            <option :value="100">100 messages</option>
            <option :value="200">200 messages</option>
            <option :value="500">500 messages</option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label">&nbsp;</label>
          <label class="form-checkbox">
            <input v-model="useLivePreview" type="checkbox" />
            <span>Use live WebSocket preview</span>
          </label>
        </div>

        <div class="form-group">
          <label class="form-label">&nbsp;</label>
          <div class="flex gap-2">
            <button class="btn btn-primary" @click="runPreview" :disabled="previewStore.loading">
              {{ previewStore.loading ? 'Loading...' : 'Run Preview' }}
            </button>
            <button
              v-if="previewStore.loading && useLivePreview"
              class="btn btn-secondary"
              @click="stopPreview"
            >
              Stop
            </button>
          </div>
        </div>
      </div>

      <div v-if="rulesStore.rules.length === 0" class="alert alert-info mt-4">
        No rules configured.
        <router-link :to="`/accounts/${accountId}/rules`">Create some rules</router-link> first.
      </div>
    </div>

    <!-- Progress -->
    <div v-if="previewStore.progress && previewStore.loading" class="card mb-4">
      <div class="progress-info">
        <span>{{ previewStore.progress.message }}</span>
        <span v-if="previewStore.progress.total > 0">
          {{ previewStore.progress.current }} / {{ previewStore.progress.total }}
        </span>
      </div>
      <div class="progress-bar mt-4">
        <div
          class="progress-bar-fill"
          :style="{
            width:
              previewStore.progress.total > 0
                ? `${(previewStore.progress.current / previewStore.progress.total) * 100}%`
                : '0%',
          }"
        ></div>
      </div>
    </div>

    <!-- Results -->
    <div v-if="previewStore.matchStats" class="card mb-4">
      <h3 class="card-title">Results</h3>
      <div class="stats-row">
        <div class="stat-box">
          <span class="stat-value">{{ previewStore.matchStats.total }}</span>
          <span class="stat-label">Total Messages</span>
        </div>
        <div class="stat-box stat-success">
          <span class="stat-value">{{ previewStore.matchStats.matched }}</span>
          <span class="stat-label">Matched</span>
        </div>
        <div class="stat-box">
          <span class="stat-value">{{ previewStore.matchStats.percentage }}%</span>
          <span class="stat-label">Match Rate</span>
        </div>
      </div>

      <div v-if="previewStore.matchStats.matched > 0" class="mt-4">
        <button class="btn btn-success" @click="applyRules" :disabled="previewStore.loading">
          Apply Rules (Move {{ previewStore.matchStats.matched }} emails)
        </button>
      </div>
    </div>

    <!-- Message List -->
    <div v-if="previewStore.messages.length > 0" class="card">
      <div class="messages-header">
        <h3 class="card-title">Messages</h3>
        <div class="filter-buttons">
          <button
            class="btn btn-sm"
            :class="filterMatched === 'all' ? 'btn-primary' : 'btn-outline'"
            @click="filterMatched = 'all'"
          >
            All ({{ previewStore.messages.length }})
          </button>
          <button
            class="btn btn-sm"
            :class="filterMatched === 'matched' ? 'btn-primary' : 'btn-outline'"
            @click="filterMatched = 'matched'"
          >
            Matched ({{ previewStore.matchedMessages.length }})
          </button>
          <button
            class="btn btn-sm"
            :class="filterMatched === 'unmatched' ? 'btn-primary' : 'btn-outline'"
            @click="filterMatched = 'unmatched'"
          >
            Unmatched ({{ previewStore.unmatchedMessages.length }})
          </button>
        </div>
      </div>

      <div class="messages-list">
        <div
          v-for="msg in displayedMessages"
          :key="msg.uid"
          class="message-item"
          :class="{ matched: msg.matched_rule }"
        >
          <div class="message-main">
            <div class="message-from">{{ msg.from }}</div>
            <div class="message-subject">{{ msg.subject || '(no subject)' }}</div>
            <div class="message-date text-muted">{{ formatDate(msg.date) }}</div>
          </div>
          <div v-if="msg.matched_rule" class="message-rule">
            <span class="badge badge-success">{{ msg.matched_rule.name }}</span>
            <span class="text-muted">&rarr; {{ msg.matched_rule.move_to_folder }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card-title {
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 1rem;
}

.preview-controls {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1rem;
  align-items: end;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  font-size: 0.875rem;
  color: var(--color-text-muted);
}

.stats-row {
  display: flex;
  gap: 2rem;
}

.stat-box {
  text-align: center;
}

.stat-box .stat-value {
  display: block;
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text);
}

.stat-box.stat-success .stat-value {
  color: var(--color-success);
}

.stat-box .stat-label {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  text-transform: uppercase;
}

.messages-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.filter-buttons {
  display: flex;
  gap: 0.5rem;
}

.messages-list {
  max-height: 600px;
  overflow-y: auto;
}

.message-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 1rem;
  border-bottom: 1px solid var(--color-border);
  gap: 1rem;
}

.message-item:last-child {
  border-bottom: none;
}

.message-item.matched {
  background: rgba(34, 197, 94, 0.05);
}

.message-main {
  flex: 1;
  min-width: 0;
}

.message-from {
  font-weight: 500;
  margin-bottom: 0.25rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.message-subject {
  font-size: 0.875rem;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.message-date {
  font-size: 0.75rem;
  margin-top: 0.25rem;
}

.message-rule {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.25rem;
  flex-shrink: 0;
}

.message-rule .text-muted {
  font-size: 0.75rem;
}
</style>
