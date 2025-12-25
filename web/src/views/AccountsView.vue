<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAccountsStore } from '../stores/accounts'
import type { AccountCreate, ConnectionStatus } from '../api/types'

const router = useRouter()
const accountsStore = useAccountsStore()

const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const testResult = ref<ConnectionStatus | null>(null)
const testing = ref(false)

const form = ref<AccountCreate>({
  name: '',
  server: '',
  port: 993,
  username: '',
  password: '',
  tls: true,
})

const sortedAccounts = computed(() => accountsStore.sortedAccounts)

function openAddModal() {
  isEditing.value = false
  editingId.value = null
  form.value = {
    name: '',
    server: '',
    port: 993,
    username: '',
    password: '',
    tls: true,
  }
  testResult.value = null
  showModal.value = true
}

function openEditModal(account: any) {
  isEditing.value = true
  editingId.value = account.id
  form.value = {
    name: account.name,
    server: account.server,
    port: account.port,
    username: account.username,
    password: '',
    tls: account.tls,
  }
  testResult.value = null
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  testResult.value = null
}

async function testConnection() {
  testing.value = true
  testResult.value = null
  try {
    if (isEditing.value && editingId.value && !form.value.password) {
      testResult.value = await accountsStore.testAccount(editingId.value)
    } else {
      testResult.value = await accountsStore.testAccountDirect(form.value)
    }
  } finally {
    testing.value = false
  }
}

async function saveAccount() {
  if (isEditing.value && editingId.value) {
    await accountsStore.updateAccount(editingId.value, form.value)
  } else {
    await accountsStore.createAccount(form.value)
  }
  closeModal()
}

async function deleteAccount(id: number) {
  if (confirm('Are you sure you want to delete this account?')) {
    await accountsStore.deleteAccount(id)
  }
}

function viewAccount(id: number) {
  router.push(`/accounts/${id}`)
}
</script>

<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">Email Accounts</h1>
      <button class="btn btn-primary" @click="openAddModal">Add Account</button>
    </div>

    <div v-if="accountsStore.error" class="alert alert-error">
      {{ accountsStore.error }}
    </div>

    <div v-if="accountsStore.loading" class="empty-state">
      <div class="loading-spinner"></div>
      <p class="mt-4">Loading accounts...</p>
    </div>

    <div v-else-if="sortedAccounts.length === 0" class="card empty-state">
      <h3>No accounts configured</h3>
      <p>Add an IMAP account to get started with email organization.</p>
      <button class="btn btn-primary mt-4" @click="openAddModal">Add Your First Account</button>
    </div>

    <div v-else class="accounts-grid">
      <div v-for="account in sortedAccounts" :key="account.id" class="card account-card">
        <div class="account-header">
          <h3>{{ account.name }}</h3>
          <div class="account-actions">
            <button class="btn btn-sm btn-outline" @click="openEditModal(account)">Edit</button>
            <button class="btn btn-sm btn-danger" @click="deleteAccount(account.id)">Delete</button>
          </div>
        </div>
        <div class="account-details">
          <p><strong>Server:</strong> {{ account.server }}:{{ account.port }}</p>
          <p><strong>Username:</strong> {{ account.username }}</p>
          <p>
            <strong>TLS:</strong>
            <span :class="account.tls ? 'badge-success' : 'badge-warning'" class="badge">{{
              account.tls ? 'Enabled' : 'Disabled'
            }}</span>
          </p>
        </div>
        <div class="account-footer">
          <button class="btn btn-primary" @click="viewAccount(account.id)">Manage</button>
        </div>
      </div>
    </div>

    <!-- Add/Edit Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h2 class="modal-title">{{ isEditing ? 'Edit Account' : 'Add Account' }}</h2>
          <button class="modal-close" @click="closeModal">&times;</button>
        </div>
        <form @submit.prevent="saveAccount" class="modal-body">
          <div class="form-group">
            <label class="form-label">Account Name</label>
            <input
              v-model="form.name"
              type="text"
              class="form-input"
              required
              placeholder="My Email"
            />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="form-group">
              <label class="form-label">IMAP Server</label>
              <input
                v-model="form.server"
                type="text"
                class="form-input"
                required
                placeholder="imap.example.com"
              />
            </div>
            <div class="form-group">
              <label class="form-label">Port</label>
              <input v-model.number="form.port" type="number" class="form-input" required />
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">Username</label>
            <input
              v-model="form.username"
              type="text"
              class="form-input"
              required
              placeholder="user@example.com"
            />
          </div>
          <div class="form-group">
            <label class="form-label"
              >Password {{ isEditing ? '(leave blank to keep current)' : '' }}</label
            >
            <input
              v-model="form.password"
              type="password"
              class="form-input"
              :required="!isEditing"
            />
          </div>
          <div class="form-group">
            <label class="form-checkbox">
              <input v-model="form.tls" type="checkbox" />
              <span>Use TLS/SSL</span>
            </label>
          </div>

          <div
            v-if="testResult"
            class="alert"
            :class="testResult.success ? 'alert-success' : 'alert-error'"
          >
            <strong>{{
              testResult.success ? 'Connection successful!' : 'Connection failed'
            }}</strong>
            <p>{{ testResult.message }}</p>
            <p v-if="testResult.success && testResult.total_emails !== undefined">
              Found {{ testResult.total_emails }} emails in INBOX
            </p>
          </div>

          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-outline"
              @click="testConnection"
              :disabled="testing"
            >
              {{ testing ? 'Testing...' : 'Test Connection' }}
            </button>
            <button type="button" class="btn btn-outline" @click="closeModal">Cancel</button>
            <button type="submit" class="btn btn-primary" :disabled="accountsStore.loading">
              {{ isEditing ? 'Save Changes' : 'Add Account' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.accounts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1rem;
}

.account-card {
  display: flex;
  flex-direction: column;
}

.account-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.account-header h3 {
  font-size: 1.125rem;
  font-weight: 600;
}

.account-actions {
  display: flex;
  gap: 0.5rem;
}

.account-details {
  flex: 1;
  font-size: 0.875rem;
}

.account-details p {
  margin-bottom: 0.5rem;
}

.account-footer {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
}
</style>
