<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useWsStore } from "../stores/wsStore";

interface PokemonInfo {
  species_id: number;
  position: number;
  current_hp: number;
  is_fainted: boolean;
}

const ws = useWsStore();

onMounted(() => {
  ws.connect();
});

const battle = computed(() => ws.battle);

const moveId = ref(1);
const switchPosition = ref(1);

const canAct = computed(() => !!ws.battleId && ws.status === "open" && !ws.battleEnded);

const switchablePokemons = computed(() => {
  const team: PokemonInfo[] = battle.value?.your_info?.team || [];
  const activePos: number = battle.value?.your_info?.active_pokemon;
  return team.filter((p) => !p.is_fainted && p.position !== activePos);
});

function onAttack() {
  ws.attack(moveId.value);
}

function onChange() {
  ws.changePokemon(switchPosition.value);
}

function clearLogs() {
  ws.logs = [];
}

function pretty(obj: unknown) {
  try { return JSON.stringify(obj, null, 2); } catch { return String(obj); }
}

function hpBarWidth(hp: number) {
  const clamped = Math.max(0, Math.min(100, Number(hp) || 0));
  return `${clamped}%`;
}

function hpBarColor(p: PokemonInfo) {
  if (p.is_fainted || p.current_hp <= 0) return "bg-zinc-600";
  if (p.current_hp <= 25) return "bg-rose-500";
  if (p.current_hp <= 50) return "bg-amber-500";
  return "bg-emerald-500";
}

function pokemonCardClass(p: PokemonInfo, activePos: number) {
  if (p.is_fainted) return "border-rose-500/30 bg-rose-950/20";
  if (p.position === activePos) return "border-blue-500/50 bg-blue-950/30";
  return "border-zinc-800 bg-zinc-950";
}

function statusBadgeClass(p: PokemonInfo, activePos: number) {
  if (p.is_fainted) return "bg-rose-500/15 text-rose-200 border border-rose-500/30";
  if (p.position === activePos) return "bg-blue-500/15 text-blue-200 border border-blue-500/30";
  return "bg-emerald-500/15 text-emerald-200 border border-emerald-500/30";
}

function statusLabel(p: PokemonInfo, activePos: number) {
  if (p.is_fainted) return "Fainted";
  if (p.position === activePos) return "Active";
  return "Alive";
}
</script>

<template>
  <div class="min-h-screen bg-zinc-950 text-zinc-100">
    <div class="mx-auto max-w-6xl px-4 py-8">
      <header class="mb-6 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold tracking-tight">Battle</h1>
          <p class="text-sm text-zinc-400">
            Status:
            <span class="ml-1 rounded-lg border border-zinc-800 bg-zinc-900 px-2 py-1 text-xs text-zinc-200">
              {{ ws.status }}
            </span>
            <span v-if="ws.battleId" class="ml-2 text-xs text-zinc-500">
              battle_id:
              <code class="rounded bg-zinc-900 px-2 py-1 text-zinc-200">{{ ws.battleId }}</code>
            </span>
          </p>
        </div>

        <!-- Action controls -->
        <div class="flex items-end gap-3">
          <!-- Attack -->
          <div class="flex flex-col gap-1">
            <label class="text-xs text-zinc-400">Move ID</label>
            <div class="flex gap-2">
              <input
                v-model.number="moveId"
                type="number"
                min="1"
                class="w-20 rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none focus:ring-2 focus:ring-indigo-500/30"
              />
              <button
                @click="onAttack"
                class="rounded-xl bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="!canAct"
              >
                Attack
              </button>
            </div>
          </div>

          <!-- Change Pokemon -->
          <div class="flex flex-col gap-1">
            <label class="text-xs text-zinc-400">Position</label>
            <div class="flex gap-2">
              <select
                v-model.number="switchPosition"
                class="w-24 rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none focus:ring-2 focus:ring-indigo-500/30"
              >
                <option
                  v-for="p in switchablePokemons"
                  :key="p.position"
                  :value="p.position"
                >
                  #{{ p.species_id }} (pos {{ p.position }})
                </option>
              </select>
              <button
                @click="onChange"
                class="rounded-xl border border-zinc-700 bg-zinc-950 px-4 py-2 text-sm font-semibold text-zinc-100 transition hover:bg-zinc-900 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="!canAct || switchablePokemons.length === 0"
              >
                Change
              </button>
            </div>
          </div>
        </div>
      </header>

      <!-- Battle message banner -->
      <div
        v-if="battle?.message"
        class="mb-4 rounded-xl border border-zinc-800 bg-zinc-900/80 px-4 py-3 text-sm text-zinc-200"
      >
        {{ battle.message }}
      </div>

      <!-- Battle ended banner -->
      <div
        v-if="ws.battleEnded && battle?.winner"
        class="mb-4 rounded-xl border px-4 py-3 text-sm font-semibold"
        :class="battle.winner === battle.your_info?.player_id
          ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-200'
          : 'border-rose-500/30 bg-rose-500/10 text-rose-200'"
      >
        {{ battle.winner === battle.your_info?.player_id ? 'You won!' : 'You lost!' }}
      </div>

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <!-- Left / main battle info -->
        <div class="lg:col-span-2">
          <div class="rounded-2xl border border-zinc-800 bg-zinc-900/60 p-5">
            <div v-if="!battle" class="text-zinc-400">
              Waiting for battle payload...
            </div>

            <div v-else class="space-y-6">
              <!-- Your info -->
              <section>
                <div class="mb-3 flex items-center justify-between">
                  <h2 class="text-lg font-semibold">You</h2>
                  <span class="text-xs text-zinc-400">
                    {{ battle.your_info?.username }} &bull;
                    Active pos: {{ battle.your_info?.active_pokemon }}
                  </span>
                </div>

                <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
                  <div
                    v-for="p in battle.your_info?.team || []"
                    :key="'you-' + p.position"
                    class="rounded-xl border p-4"
                    :class="pokemonCardClass(p, battle.your_info?.active_pokemon)"
                  >
                    <div class="flex items-center justify-between">
                      <div class="text-sm font-semibold">Species #{{ p.species_id }}</div>
                      <div
                        class="rounded-md px-2 py-1 text-xs font-medium"
                        :class="statusBadgeClass(p, battle.your_info?.active_pokemon)"
                      >
                        {{ statusLabel(p, battle.your_info?.active_pokemon) }}
                      </div>
                    </div>

                    <div class="mt-3 text-xs text-zinc-400">Position: {{ p.position }}</div>

                    <div class="mt-2">
                      <div class="flex items-center justify-between text-xs text-zinc-400">
                        <span>HP</span>
                        <span class="text-zinc-200">{{ p.current_hp }}</span>
                      </div>
                      <div class="mt-2 h-2 w-full overflow-hidden rounded-full bg-zinc-800">
                        <div
                          class="h-2 rounded-full transition-all duration-300"
                          :class="hpBarColor(p)"
                          :style="{ width: hpBarWidth(p.current_hp) }"
                        ></div>
                      </div>
                    </div>
                  </div>
                </div>
              </section>

              <!-- Opponent info -->
              <section>
                <div class="mb-3 flex items-center justify-between">
                  <h2 class="text-lg font-semibold">Opponent</h2>
                  <span class="text-xs text-zinc-400">
                    {{ battle.opponent_info?.username }} &bull;
                    Active pos: {{ battle.opponent_info?.active_pokemon }}
                  </span>
                </div>

                <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
                  <div
                    v-for="p in battle.opponent_info?.team || []"
                    :key="'opp-' + p.position"
                    class="rounded-xl border p-4"
                    :class="pokemonCardClass(p, battle.opponent_info?.active_pokemon)"
                  >
                    <div class="flex items-center justify-between">
                      <div class="text-sm font-semibold">Species #{{ p.species_id }}</div>
                      <div
                        class="rounded-md px-2 py-1 text-xs font-medium"
                        :class="statusBadgeClass(p, battle.opponent_info?.active_pokemon)"
                      >
                        {{ statusLabel(p, battle.opponent_info?.active_pokemon) }}
                      </div>
                    </div>

                    <div class="mt-3 text-xs text-zinc-400">Position: {{ p.position }}</div>

                    <div class="mt-2">
                      <div class="flex items-center justify-between text-xs text-zinc-400">
                        <span>HP</span>
                        <span class="text-zinc-200">{{ p.current_hp }}</span>
                      </div>
                      <div class="mt-2 h-2 w-full overflow-hidden rounded-full bg-zinc-800">
                        <div
                          class="h-2 rounded-full transition-all duration-300"
                          :class="hpBarColor(p)"
                          :style="{ width: hpBarWidth(p.current_hp) }"
                        ></div>
                      </div>
                    </div>
                  </div>
                </div>
              </section>

              <!-- Raw payload -->
              <details class="rounded-xl border border-zinc-800 bg-zinc-950 p-4">
                <summary class="cursor-pointer text-sm text-zinc-300">Raw payload</summary>
                <pre class="mt-3 max-h-72 overflow-auto text-xs text-zinc-200">{{ pretty(battle) }}</pre>
              </details>
            </div>
          </div>
        </div>

        <!-- Right / logs -->
        <div class="lg:col-span-1">
          <div class="rounded-2xl border border-zinc-800 bg-zinc-900/60 p-5">
            <div class="mb-3 flex items-center justify-between">
              <h2 class="text-lg font-semibold">Battle logs</h2>
              <button
                class="text-xs text-zinc-400 hover:text-zinc-200"
                @click="clearLogs"
              >
                Clear
              </button>
            </div>

            <div class="h-[60vh] overflow-auto rounded-xl border border-zinc-800 bg-zinc-950 p-3">
              <div v-if="!ws.logs.length" class="text-sm text-zinc-500">
                No logs yet...
              </div>

              <ul class="space-y-2">
                <li
                  v-for="(line, idx) in ws.logs"
                  :key="idx"
                  class="text-xs text-zinc-200"
                >
                  {{ line }}
                </li>
              </ul>
            </div>

            <p class="mt-3 text-xs text-zinc-500">
              Logs append on server events (MatchFound, Attack, Change, etc.).
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>