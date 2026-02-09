import { defineStore } from "pinia";
import { getSocket, setSocket } from "../ws/socket";

const WS_URL = import.meta.env.VITE_WS_URL || "ws://localhost:3003";

// Numeric message types (must match server)
export const CLIENT_MESSAGE_TYPE = {
  Connect: 1,
  Attack: 2,
  ChangePokemon: 3,
  Surrender: 4,
  Status: 5,
  Match: 6,
};

export const SERVER_MESSAGE_TYPE = {
  AcceptConnection: 50,
  Attack: 51,
  ChangePokemon: 52,
  Status: 53,
  BattleEnded: 54,
  Disconnect: 55,
  Error: 56,
  MatchFound: 57,
  QueueJoined: 58,
};

function createMessage(type, payload) {
  return { type, payload };
}

export const useWsStore = defineStore("ws", {
  state: () => ({
    status: "idle", // idle | connecting | open | closed | error

    username: "",
    pokemons: [],

    accepted: false,
    queueJoined: false,

    // battle state
    battleId: null, // string
    battle: null, // object (match found payload or status payload)
    battleEnded: false,

    // logs panel
    logs: [], // string[]

    lastServerMessage: null,
    messages: [],
  }),

  actions: {
    addLog(line) {
      const ts = new Date().toLocaleTimeString();
      this.logs.push(`[${ts}] ${line}`);
      if (this.logs.length > 200) this.logs.shift();
    },

    connect() {
      const existing = getSocket();
      if (
        existing &&
        (existing.readyState === WebSocket.OPEN ||
          existing.readyState === WebSocket.CONNECTING)
      ) {
        return existing;
      }

      this.status = "connecting";
      this.accepted = false;
      this.queueJoined = false;

      const ws = new WebSocket(WS_URL);
      setSocket(ws);

      ws.onopen = () => {
        this.status = "open";
      };

      ws.onclose = () => {
        this.status = "closed";
        this.accepted = false;
        this.queueJoined = false;
      };

      ws.onerror = () => {
        this.status = "error";
      };

      ws.onmessage = (event) => {
        let msg = event.data;
        try {
          msg = JSON.parse(event.data);
        } catch {
          // if server ever sends plain text, keep it as-is
        }

        this.lastServerMessage = msg;
        this.messages.push(msg);

        switch (msg?.type) {
          case SERVER_MESSAGE_TYPE.AcceptConnection:
            this.accepted = true;
            this.addLog("Connected (AcceptConnection).");
            break;

          case SERVER_MESSAGE_TYPE.QueueJoined:
            this.queueJoined = true;
            this.addLog("Joined matchmaking queue.");
            break;

          case SERVER_MESSAGE_TYPE.MatchFound:
            // payload has battle_id + player info
            this.battleId = msg.payload?.battle_id ?? null;
            this.battle = msg.payload;
            this.addLog(`Match found. battle_id=${this.battleId}`);
            break;

          case SERVER_MESSAGE_TYPE.Status:
            // server battle updates
            this.battle = msg.payload;
            // if status payload also has battle_id, keep it synced
            if (msg.payload?.battle_id) this.battleId = msg.payload.battle_id;
            this.addLog("Battle status update received.");
            break;

          case SERVER_MESSAGE_TYPE.Attack:
            this.battle = msg.payload;
            this.addLog(`Attack: ${msg.payload?.message || safeString(msg.payload)}`);
            break;

          case SERVER_MESSAGE_TYPE.ChangePokemon:
            this.battle = msg.payload;
            this.addLog(`Change: ${msg.payload?.message || safeString(msg.payload)}`);
            break;

          case SERVER_MESSAGE_TYPE.BattleEnded:
            this.battle = msg.payload;
            this.battleEnded = true;
            this.addLog(`Battle ended! Winner: ${msg.payload?.winner || "unknown"}`);
            break;

          case SERVER_MESSAGE_TYPE.Disconnect:
            this.addLog("Server disconnected you.");
            break;

          case SERVER_MESSAGE_TYPE.Error:
            this.addLog(`Server error: ${safeString(msg.payload)}`);
            break;

          default:
            // Uncomment if you want to log every unknown message:
            // this.addLog(`Server msg type=${msg?.type}: ${safeString(msg?.payload)}`);
            break;
        }
      };

      return ws;
    },

    disconnect() {
      const ws = getSocket();
      if (ws) ws.close();
      setSocket(null);

      this.status = "closed";
      this.accepted = false;
      this.queueJoined = false;
    },

    sendMessage(type, payload) {
      const ws = getSocket();
      if (!ws || ws.readyState !== WebSocket.OPEN) {
        throw new Error("WebSocket is not open");
      }
      ws.send(JSON.stringify(createMessage(type, payload)));
    },

    waitUntilOpen(timeoutMs = 3000) {
      const ws = this.connect();

      return new Promise((resolve, reject) => {
        if (ws.readyState === WebSocket.OPEN) return resolve(ws);

        const t = setTimeout(() => {
          cleanup();
          reject(new Error("Timed out waiting for WebSocket to open"));
        }, timeoutMs);

        const onOpen = () => {
          cleanup();
          resolve(ws);
        };

        const onErr = () => {
          cleanup();
          reject(new Error("WebSocket error while opening"));
        };

        const cleanup = () => {
          clearTimeout(t);
          ws.removeEventListener("open", onOpen);
          ws.removeEventListener("error", onErr);
        };

        ws.addEventListener("open", onOpen);
        ws.addEventListener("error", onErr);
      });
    },

    waitForServerType(expectedType, timeoutMs = 3000) {
      return new Promise((resolve, reject) => {
        const ws = getSocket();
        if (!ws) return reject(new Error("No WebSocket instance"));

        const t = setTimeout(() => {
          cleanup();
          reject(
            new Error(`Timed out waiting for server message type ${expectedType}`)
          );
        }, timeoutMs);

        const onMessage = (event) => {
          let msg = event.data;
          try {
            msg = JSON.parse(event.data);
          } catch {
            return;
          }

          if (msg?.type === expectedType) {
            cleanup();
            resolve(msg);
          }
        };

        const cleanup = () => {
          clearTimeout(t);
          ws.removeEventListener("message", onMessage);
        };

        ws.addEventListener("message", onMessage);
      });
    },

    async connectAndAccept(username, pokemons) {
      this.username = username;
      this.pokemons = pokemons;

      await this.waitUntilOpen();

      this.sendMessage(CLIENT_MESSAGE_TYPE.Connect, { username, pokemons });

      const accept = await this.waitForServerType(
        SERVER_MESSAGE_TYPE.AcceptConnection,
        3000
      );
      this.accepted = true;

      return accept;
    },

    joinQueue() {
      this.sendMessage(CLIENT_MESSAGE_TYPE.Match, {});
      this.addLog("Sent Match (join queue) request.");
    },

    // Actions that include battle_id (since your server needs it)
    attack(moveId = 1) {
      if (!this.battleId) throw new Error("No battleId yet");
      this.sendMessage(CLIENT_MESSAGE_TYPE.Attack, {
        battle_id: this.battleId,
        move_id: moveId,
      });
      this.addLog(`You attacked (move_id=${moveId}).`);
    },

    changePokemon(position = 2) {
      if (!this.battleId) throw new Error("No battleId yet");
      this.sendMessage(CLIENT_MESSAGE_TYPE.ChangePokemon, {
        battle_id: this.battleId,
        position,
      });
      this.addLog(`You changed Pokémon (position=${position}).`);
    },

    surrender() {
      if (!this.battleId) throw new Error("No battleId yet");
      this.sendMessage(CLIENT_MESSAGE_TYPE.Surrender, {
        battle_id: this.battleId,
      });
      this.addLog("You surrendered.");
    },

    requestStatus() {
      if (!this.battleId) throw new Error("No battleId yet");
      this.sendMessage(CLIENT_MESSAGE_TYPE.Status, {
        battle_id: this.battleId,
      });
      this.addLog("Requested battle status.");
    },
  },
});

function safeString(v) {
  try {
    return JSON.stringify(v);
  } catch {
    return String(v);
  }
}