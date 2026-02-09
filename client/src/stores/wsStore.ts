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
    match: null,
    battle: null,

    lastServerMessage: null,
    messages: [],
  }),

  actions: {
    connect() {
      const existing = getSocket();
      if (existing && (existing.readyState === WebSocket.OPEN || existing.readyState === WebSocket.CONNECTING)) {
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
          // if server sends non-json, keep raw
        }

        this.lastServerMessage = msg;
        this.messages.push(msg);

        // Route by numeric type
        switch (msg?.type) {
          case SERVER_MESSAGE_TYPE.AcceptConnection:
            this.accepted = true;
            break;

          case SERVER_MESSAGE_TYPE.QueueJoined:
            this.queueJoined = true;
            break;

          case SERVER_MESSAGE_TYPE.MatchFound:
            this.match = msg.payload;
            break;

          case SERVER_MESSAGE_TYPE.Status:
            this.battle = msg.payload;
            break;

          case SERVER_MESSAGE_TYPE.BattleEnded:
            // optional: store end info
            // this.battleEnded = msg.payload;
            break;

          case SERVER_MESSAGE_TYPE.Error:
            // optional: store server error payload
            // this.serverError = msg.payload;
            break;

          default:
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
          reject(new Error(`Timed out waiting for server message type ${expectedType}`));
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

    // Matches your integration test exactly:
    // connect -> send Connect(1) -> wait for AcceptConnection(50)
    async connectAndAccept(username, pokemons) {
      this.username = username;
      this.pokemons = pokemons;

      await this.waitUntilOpen();

      this.sendMessage(CLIENT_MESSAGE_TYPE.Connect, { username, pokemons });

      const accept = await this.waitForServerType(SERVER_MESSAGE_TYPE.AcceptConnection, 3000);
      this.accepted = true;

      return accept;
    },

    // Optional: if your server requires you to explicitly join matchmaking queue:
    joinQueue() {
      this.sendMessage(CLIENT_MESSAGE_TYPE.Match, {});
    },

    // Battle actions
    attack(moveIndex) {
      this.sendMessage(CLIENT_MESSAGE_TYPE.Attack, { moveIndex });
    },

    changePokemon(pokemonIndex) {
      this.sendMessage(CLIENT_MESSAGE_TYPE.ChangePokemon, { pokemonIndex });
    },

    surrender() {
      this.sendMessage(CLIENT_MESSAGE_TYPE.Surrender, {});
    },

    requestStatus() {
      this.sendMessage(CLIENT_MESSAGE_TYPE.Status, {});
    },
  },
});