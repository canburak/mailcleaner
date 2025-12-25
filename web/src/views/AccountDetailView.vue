<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAccountsStore } from '../stores/accounts'
import { useRulesStore } from '../stores/rules'

const props = defineProps<{ id: string }>()
const route = useRoute()
const router = useRouter()
const accountsStore = useAccountsStore()
const rulesStore = useRulesStore()

const accountId = ref(parseInt(props.id))

onMounted(async () => {
  await accountsStore.fetchAccount(accountId.value)
  await accountsStore.fetchFolders(accountId.value)
  await rulesStore.fetchRules(accountId.value)
})

watch(
  () => props.id,
  async newId => {
    accountId.value = parseInt(newId)
    await accountsStore.fetchAccount(accountId.value)
    await accountsStore.fetchFolders(accountId.value)
    await rulesStore.fetchRules(accountId.value)
  }
)

function goToRules() {
  router.push(`/accounts/${accountId.value}/rules`)
}

function goToPreview() {
  router.push(`/accounts/${accountId.value}/preview`)
}

function goBack() {
  router.push('/accounts')
}
</script>

<template>
  <div>
    <div class="page-header">
      <div class="flex items-center gap-4">
        <button class="btn btn-outline" @click="goBack">&larr; Back</button>
        <h1 class="page-title">{{ accountsStore.currentAccount?.name || 'Account Details' }}</h1>
      </div>
    </div>

    <div v-if="accountsStore.loading" class="empty-state">
      <div class="loading-spinner"></div>
      <p class="mt-4">Loading account...</p>
    </div>

    <div v-else-if="!accountsStore.currentAccount" class="card empty-state">
      <h3>Account not found</h3>
      <button class="btn btn-primary mt-4" @click="goBack">Back to Accounts</button>
    </div>

    <template v-else>
      <div class="grid grid-cols-2 gap-4 mb-4">
        <div class="card">
          <h3 class="card-title">Account Information</h3>
          <div class="info-list">
            <div class="info-item">
              <span class="info-label">Server</span>
              <span class="info-value"
                >{{ accountsStore.currentAccount.server }}:{{
                  accountsStore.currentAccount.port
                }}</span
              >
            </div>
            <div class="info-item">
              <span class="info-label">Username</span>
              <span class="info-value">{{ accountsStore.currentAccount.username }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">TLS</span>
              <span class="info-value">
                <span
                  :class="accountsStore.currentAccount.tls ? 'badge-success' : 'badge-warning'"
                  class="badge"
                >
                  {{ accountsStore.currentAccount.tls ? 'Enabled' : 'Disabled' }}
                </span>
              </span>
            </div>
          </div>
        </div>

        <div class="card">
          <h3 class="card-title">Quick Stats</h3>
          <div class="stats-grid">
            <div class="stat-item">
              <span class="stat-value">{{ rulesStore.rules.length }}</span>
              <span class="stat-label">Rules</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ rulesStore.enabledRules.length }}</span>
              <span class="stat-label">Active</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ accountsStore.folders.length }}</span>
              <span class="stat-label">Folders</span>
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div class="card action-card" @click="goToRules">
          <div class="action-icon">📋</div>
          <h3>Manage Rules</h3>
          <p>Create, edit, and organize email filtering rules</p>
        </div>

        <div class="card action-card" @click="goToPreview">
          <div class="action-icon">👁️</div>
          <h3>Live Preview</h3>
          <p>Test rules against live emails and see matches</p>
        </div>
      </div>

      <div class="card mt-4">
        <h3 class="card-title">Folders</h3>
        <div v-if="accountsStore.folders.length === 0" class="text-muted">
          No folders loaded. Click "Live Preview" to connect and load folders.
        </div>
        <div v-else class="folders-list">
          <div v-for="folder in accountsStore.folders" :key="folder.name" class="folder-item">
            <span class="folder-name">{{ folder.name }}</span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.card-title {
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 1rem;
  color: var(--color-text);
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--color-border);
}

.info-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.info-label {
  color: var(--color-text-muted);
  font-size: 0.875rem;
}

.info-value {
  font-weight: 500;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  text-align: center;
}

.stat-item {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-primary);
}

.stat-label {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  text-transform: uppercase;
}

.action-card {
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}

.action-card:hover {
  border-color: var(--color-primary);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.15);
}

.action-icon {
  font-size: 2.5rem;
  margin-bottom: 0.5rem;
}

.action-card h3 {
  font-size: 1.125rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.action-card p {
  color: var(--color-text-muted);
  font-size: 0.875rem;
}

.folders-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.folder-item {
  background: var(--color-bg);
  padding: 0.5rem 1rem;
  border-radius: var(--radius);
  font-size: 0.875rem;
}
</style>
