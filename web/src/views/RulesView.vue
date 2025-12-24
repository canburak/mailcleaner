<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAccountsStore } from '../stores/accounts';
import { useRulesStore } from '../stores/rules';
import type { RuleCreate } from '../api/types';

const props = defineProps<{ id: string }>();
const router = useRouter();
const accountsStore = useAccountsStore();
const rulesStore = useRulesStore();

const accountId = ref(parseInt(props.id));
const showModal = ref(false);
const isEditing = ref(false);
const editingId = ref<number | null>(null);

const form = ref<RuleCreate>({
  name: '',
  pattern: '',
  pattern_type: 'sender',
  move_to_folder: '',
  enabled: true,
  priority: 0,
});

onMounted(async () => {
  await accountsStore.fetchAccount(accountId.value);
  await accountsStore.fetchFolders(accountId.value);
  await rulesStore.fetchRules(accountId.value);
});

watch(() => props.id, async (newId) => {
  accountId.value = parseInt(newId);
  await accountsStore.fetchAccount(accountId.value);
  await accountsStore.fetchFolders(accountId.value);
  await rulesStore.fetchRules(accountId.value);
});

function goBack() {
  router.push(`/accounts/${accountId.value}`);
}

function openAddModal() {
  isEditing.value = false;
  editingId.value = null;
  form.value = {
    name: '',
    pattern: '',
    pattern_type: 'sender',
    move_to_folder: '',
    enabled: true,
    priority: rulesStore.rules.length,
  };
  showModal.value = true;
}

function openEditModal(rule: any) {
  isEditing.value = true;
  editingId.value = rule.id;
  form.value = {
    name: rule.name,
    pattern: rule.pattern,
    pattern_type: rule.pattern_type,
    move_to_folder: rule.move_to_folder,
    enabled: rule.enabled,
    priority: rule.priority,
  };
  showModal.value = true;
}

function closeModal() {
  showModal.value = false;
}

async function saveRule() {
  if (isEditing.value && editingId.value) {
    await rulesStore.updateRule(editingId.value, form.value);
  } else {
    await rulesStore.createRule(accountId.value, form.value);
  }
  closeModal();
}

async function deleteRule(id: number) {
  if (confirm('Are you sure you want to delete this rule?')) {
    await rulesStore.deleteRule(id);
  }
}

async function toggleRule(id: number) {
  await rulesStore.toggleRule(id);
}
</script>

<template>
  <div>
    <div class="page-header">
      <div class="flex items-center gap-4">
        <button class="btn btn-outline" @click="goBack">&larr; Back</button>
        <h1 class="page-title">Rules for {{ accountsStore.currentAccount?.name }}</h1>
      </div>
      <button class="btn btn-primary" @click="openAddModal">Add Rule</button>
    </div>

    <div v-if="rulesStore.error" class="alert alert-error">
      {{ rulesStore.error }}
    </div>

    <div v-if="rulesStore.loading" class="empty-state">
      <div class="loading-spinner"></div>
      <p class="mt-4">Loading rules...</p>
    </div>

    <div v-else-if="rulesStore.rules.length === 0" class="card empty-state">
      <h3>No rules configured</h3>
      <p>Create rules to automatically organize your emails.</p>
      <button class="btn btn-primary mt-4" @click="openAddModal">Create First Rule</button>
    </div>

    <div v-else class="rules-list">
      <div v-for="rule in rulesStore.sortedRules" :key="rule.id" class="card rule-card" :class="{ disabled: !rule.enabled }">
        <div class="rule-header">
          <div class="rule-info">
            <label class="form-checkbox">
              <input type="checkbox" :checked="rule.enabled" @change="toggleRule(rule.id)" />
            </label>
            <div>
              <h3>{{ rule.name }}</h3>
              <span class="badge badge-info">Priority: {{ rule.priority }}</span>
            </div>
          </div>
          <div class="rule-actions">
            <button class="btn btn-sm btn-outline" @click="openEditModal(rule)">Edit</button>
            <button class="btn btn-sm btn-danger" @click="deleteRule(rule.id)">Delete</button>
          </div>
        </div>
        <div class="rule-details">
          <div class="rule-condition">
            <span class="label">When</span>
            <span class="badge badge-info">{{ rule.pattern_type }}</span>
            <span class="label">contains</span>
            <code>{{ rule.pattern }}</code>
          </div>
          <div class="rule-action">
            <span class="label">Move to</span>
            <code>{{ rule.move_to_folder }}</code>
          </div>
        </div>
      </div>
    </div>

    <!-- Add/Edit Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h2 class="modal-title">{{ isEditing ? 'Edit Rule' : 'Add Rule' }}</h2>
          <button class="modal-close" @click="closeModal">&times;</button>
        </div>
        <form @submit.prevent="saveRule" class="modal-body">
          <div class="form-group">
            <label class="form-label">Rule Name</label>
            <input v-model="form.name" type="text" class="form-input" required placeholder="Newsletter Filter" />
          </div>

          <div class="form-group">
            <label class="form-label">Pattern Type</label>
            <select v-model="form.pattern_type" class="form-select" required>
              <option value="sender">Sender (From address)</option>
              <option value="subject">Subject line</option>
              <option value="from_domain">Sender domain</option>
            </select>
          </div>

          <div class="form-group">
            <label class="form-label">Pattern</label>
            <input v-model="form.pattern" type="text" class="form-input" required placeholder="newsletter@, github.com" />
            <small class="text-muted">The text to match (case-insensitive)</small>
          </div>

          <div class="form-group">
            <label class="form-label">Move to Folder</label>
            <input v-model="form.move_to_folder" type="text" class="form-input" list="folders" required placeholder="Newsletters" />
            <datalist id="folders">
              <option v-for="folder in accountsStore.folders" :key="folder.name" :value="folder.name">
                {{ folder.name }}
              </option>
            </datalist>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="form-group">
              <label class="form-label">Priority</label>
              <input v-model.number="form.priority" type="number" class="form-input" min="0" />
              <small class="text-muted">Higher priority rules are checked first</small>
            </div>

            <div class="form-group">
              <label class="form-label">&nbsp;</label>
              <label class="form-checkbox">
                <input v-model="form.enabled" type="checkbox" />
                <span>Rule enabled</span>
              </label>
            </div>
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-outline" @click="closeModal">Cancel</button>
            <button type="submit" class="btn btn-primary" :disabled="rulesStore.loading">
              {{ isEditing ? 'Save Changes' : 'Create Rule' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.rules-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.rule-card {
  transition: opacity 0.2s;
}

.rule-card.disabled {
  opacity: 0.6;
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.rule-info {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}

.rule-info h3 {
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.rule-actions {
  display: flex;
  gap: 0.5rem;
}

.rule-details {
  background: var(--color-bg);
  border-radius: var(--radius);
  padding: 1rem;
}

.rule-condition,
.rule-action {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.rule-condition {
  margin-bottom: 0.5rem;
}

.rule-details .label {
  color: var(--color-text-muted);
  font-size: 0.875rem;
}

.rule-details code {
  background: var(--color-surface);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.875rem;
  border: 1px solid var(--color-border);
}
</style>
