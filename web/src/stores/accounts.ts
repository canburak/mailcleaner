import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { accountsApi } from '../api/client';
import type { Account, AccountCreate, ConnectionStatus, Folder } from '../api/types';

export const useAccountsStore = defineStore('accounts', () => {
  const accounts = ref<Account[]>([]);
  const currentAccount = ref<Account | null>(null);
  const folders = ref<Folder[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const sortedAccounts = computed(() =>
    [...accounts.value].sort((a, b) => a.name.localeCompare(b.name))
  );

  async function fetchAccounts() {
    loading.value = true;
    error.value = null;
    try {
      accounts.value = await accountsApi.list();
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message;
    } finally {
      loading.value = false;
    }
  }

  async function fetchAccount(id: number) {
    loading.value = true;
    error.value = null;
    try {
      currentAccount.value = await accountsApi.get(id);
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message;
    } finally {
      loading.value = false;
    }
  }

  async function createAccount(data: AccountCreate): Promise<Account | null> {
    loading.value = true;
    error.value = null;
    try {
      const account = await accountsApi.create(data);
      accounts.value.push(account);
      return account;
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message;
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function updateAccount(id: number, data: Partial<AccountCreate>): Promise<Account | null> {
    loading.value = true;
    error.value = null;
    try {
      const account = await accountsApi.update(id, data);
      const index = accounts.value.findIndex(a => a.id === id);
      if (index >= 0) {
        accounts.value[index] = account;
      }
      if (currentAccount.value?.id === id) {
        currentAccount.value = account;
      }
      return account;
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message;
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function deleteAccount(id: number): Promise<boolean> {
    loading.value = true;
    error.value = null;
    try {
      await accountsApi.delete(id);
      accounts.value = accounts.value.filter(a => a.id !== id);
      if (currentAccount.value?.id === id) {
        currentAccount.value = null;
      }
      return true;
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message;
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function testAccount(id: number): Promise<ConnectionStatus | null> {
    loading.value = true;
    error.value = null;
    try {
      return await accountsApi.test(id);
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message;
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function testAccountDirect(data: AccountCreate): Promise<ConnectionStatus | null> {
    loading.value = true;
    error.value = null;
    try {
      return await accountsApi.testDirect(data);
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message;
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function fetchFolders(id: number) {
    try {
      folders.value = await accountsApi.getFolders(id);
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message;
    }
  }

  async function createFolder(id: number, name: string): Promise<boolean> {
    try {
      await accountsApi.createFolder(id, name);
      await fetchFolders(id);
      return true;
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message;
      return false;
    }
  }

  function selectAccount(account: Account | null) {
    currentAccount.value = account;
    if (account) {
      fetchFolders(account.id);
    } else {
      folders.value = [];
    }
  }

  return {
    accounts,
    currentAccount,
    folders,
    loading,
    error,
    sortedAccounts,
    fetchAccounts,
    fetchAccount,
    createAccount,
    updateAccount,
    deleteAccount,
    testAccount,
    testAccountDirect,
    fetchFolders,
    createFolder,
    selectAccount,
  };
});
