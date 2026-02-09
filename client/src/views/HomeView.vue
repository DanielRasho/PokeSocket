<template>
  <div class="w-full bg-zinc-950 text-zinc-100 flex justify-center">
    <div class="mx-auto max-w-2xl px-4 py-10">
      <header class="mb-8">
        <h1 class="text-2xl font-semibold tracking-tight">Register & Matchmaking</h1>
        <p class="mt-1 text-sm text-zinc-400">
          Enter a username and choose 3 Pokémon IDs to join the queue.
        </p>
      </header>

      <!-- Register Card -->
      <div class="rounded-2xl border border-zinc-800 bg-zinc-900/60 p-5 shadow-sm">
        <div class="space-y-5">
          <!-- Username -->
          <div>
            <label class="mb-2 block text-sm font-medium text-zinc-200">Username</label>
            <input
              v-model.trim="username"
              type="text"
              placeholder="AshKetchum"
              autocomplete="off"
              class="w-full rounded-xl border border-zinc-800 bg-zinc-950 px-4 py-3 text-zinc-100 placeholder:text-zinc-600 outline-none ring-0 transition focus:border-zinc-700 focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
            />
          </div>

          <!-- Pokemon IDs -->
          <div>
            <label class="mb-2 block text-sm font-medium text-zinc-200">
              Pick 3 Pokémon (integer IDs)
            </label>

            <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
              <input
                v-model.number="pokemon1"
                type="number"
                min="1"
                step="1"
                class="w-full rounded-xl border border-zinc-800 bg-zinc-950 px-4 py-3 text-zinc-100 placeholder:text-zinc-600 outline-none transition focus:border-zinc-700 focus:ring-2 focus:ring-indigo-500/30"
              />
              <input
                v-model.number="pokemon2"
                type="number"
                min="1"
                step="1"
                class="w-full rounded-xl border border-zinc-800 bg-zinc-950 px-4 py-3 text-zinc-100 placeholder:text-zinc-600 outline-none transition focus:border-zinc-700 focus:ring-2 focus:ring-indigo-500/30"
              />
              <input
                v-model.number="pokemon3"
                type="number"
                min="1"
                step="1"
                class="w-full rounded-xl border border-zinc-800 bg-zinc-950 px-4 py-3 text-zinc-100 placeholder:text-zinc-600 outline-none transition focus:border-zinc-700 focus:ring-2 focus:ring-indigo-500/30"
              />
            </div>

            <p v-if="pokemonError" class="mt-2 text-sm text-rose-400">
              {{ pokemonError }}
            </p>
          </div>

          <!-- Actions -->
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <button
              :disabled="!canRegister || loading"
              @click="onRegister"
              class="inline-flex items-center justify-center rounded-xl bg-indigo-600 px-5 py-3 text-sm font-semibold text-white transition hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <span v-if="!loading">Register</span>
              <span v-else class="inline-flex items-center gap-2">
                <span
                  class="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white"
                ></span>
                Registering...
              </span>
            </button>

            <div class="text-sm text-zinc-400">
              Status:
              <span class="ml-1 rounded-lg px-2 py-1 text-xs font-medium" :class="statusPillClass">
                {{ ws.status }}
              </span>
            </div>
          </div>

          <!-- Debug WS URL -->
          <div class="text-xs text-zinc-500">
            WebSocket URL:
            <code class="rounded bg-zinc-950 px-2 py-1 text-zinc-300">
              {{ wsUrlShown }}
            </code>
          </div>

          <!-- Error box -->
          <div
            v-if="error"
            class="rounded-xl border border-rose-900/60 bg-rose-950/40 px-4 py-3 text-sm text-rose-200"
          >
            {{ error }}
          </div>
        </div>
      </div>

      <!-- Match found -->
      <div
        v-if="ws.battleId && ws.battle"
        class="mt-6 rounded-2xl border border-zinc-800 bg-zinc-900/60 p-5"
      >
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="text-lg font-semibold">Match found!</h2>
            <p class="text-sm text-zinc-400">
              battle_id:
              <code class="rounded bg-zinc-950 px-2 py-1 text-zinc-200">{{ ws.battleId }}</code>
            </p>
          </div>
          <button
            @click="goBattle"
            class="rounded-xl border border-zinc-700 bg-zinc-950 px-5 py-3 text-sm font-semibold text-zinc-100 transition hover:bg-zinc-900"
          >
            Go to battle
          </button>
        </div>

        <pre
          class="mt-4 max-h-64 overflow-auto rounded-xl border border-zinc-800 bg-zinc-950 p-4 text-xs text-zinc-200"
          >{{ pretty(ws.battle) }}</pre
        >
      </div>

      <!-- Last messages -->
      <div
        v-if="ws.messages?.length"
        class="mt-6 rounded-2xl border border-zinc-800 bg-zinc-900/60 p-5"
      >
        <h2 class="text-lg font-semibold">Last messages</h2>
        <p class="text-sm text-zinc-400">Showing the last 5 received.</p>

        <pre
          class="mt-4 max-h-64 overflow-auto rounded-xl border border-zinc-800 bg-zinc-950 p-4 text-xs text-zinc-200"
          >{{ lastMessages }}</pre
        >
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useWsStore } from '../stores/wsStore'

const ws = useWsStore()
const router = useRouter()

const username = ref('')
const pokemon1 = ref(1)
const pokemon2 = ref(2)
const pokemon3 = ref(3)

const loading = ref(false)
const error = ref('')

const wsUrlShown = computed(() => import.meta.env.VITE_WS_URL || 'ws://localhost:3003')

const pokemonList = computed(() => [pokemon1.value, pokemon2.value, pokemon3.value])

const pokemonError = computed(() => {
  const arr = pokemonList.value

  if (!arr.every((n) => Number.isInteger(n))) return 'All Pokémon IDs must be integers.'
  if (!arr.every((n) => n > 0)) return 'Pokémon IDs must be positive numbers.'

  // Optional: prevent duplicates
  const uniq = new Set(arr)
  if (uniq.size !== 3) return 'Pick 3 different Pokémon IDs (no duplicates).'

  return ''
})

const canRegister = computed(() => username.value.trim().length > 0 && pokemonError.value === '')

watch([username, pokemon1, pokemon2, pokemon3], () => {
  error.value = ''
})

watch(
  () => ws.battleId,
  (id) => {
    if (id && router.currentRoute.value.name !== "battle") {
      router.push({ name: "battle" });
    }
  },
  { immediate: true }
);

const statusPillClass = computed(() => {
  switch (ws.status) {
    case 'open':
      return 'bg-emerald-500/15 text-emerald-200 border border-emerald-500/30'
    case 'connecting':
      return 'bg-amber-500/15 text-amber-200 border border-amber-500/30'
    case 'error':
      return 'bg-rose-500/15 text-rose-200 border border-rose-500/30'
    case 'closed':
      return 'bg-zinc-800 text-zinc-200 border border-zinc-700'
    default:
      return 'bg-zinc-800 text-zinc-200 border border-zinc-700'
  }
})

async function onRegister() {
  if (!canRegister.value) return;

  loading.value = true;
  error.value = "";

  // clear old battle so the watcher doesn't immediately redirect
  ws.battleId = null;
  ws.battle = null;

  try {
    await ws.connectAndAccept(username.value.trim(), pokemonList.value);
    ws.joinQueue();
  } catch (e) {
    error.value = e?.message || "Failed to register.";
  } finally {
    loading.value = false;
  }
}

function goBattle() {
  router.push({ name: 'battle' })
}

function pretty(obj) {
  try {
    return JSON.stringify(obj, null, 2)
  } catch {
    return String(obj)
  }
}

const lastMessages = computed(() => {
  const msgs = ws.messages || []
  return JSON.stringify(msgs.slice(-5), null, 2)
})
</script>
