import axios from 'axios';
import WebSocket from 'ws';
import { number, object, string } from 'yup';

export const WS_URL = 'ws://localhost:3003/battle';

// Message types matching your Go server
export const CLIENT_MESSAGE_TYPE = {
  Connect: 1,
  Attack: 2,
  ChangePokemon: 3,
  Surrender: 4,
  Status: 5,
  Match: 6,
} as const;

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
} as const;

export interface Message<T = any> {
  type: number;
  payload: T;
}

// =======================
// SCHEMAS 
// =======================

export const CONNECT_REQUEST = (username: string, pokemons: number[]) => {
    return createMessage(CLIENT_MESSAGE_TYPE.Connect, {
      username: username,
      pokemons: pokemons,
    });
  };

export const CONNECT_SCHEMA = object().shape({
  id: string().uuid().required(),
  username: string().required(),
});

export const MATCH_REQUEST = () => {
  return createMessage(CLIENT_MESSAGE_TYPE.Match, {});
};

export const MATCH_FOUND_SCHEMA = object().shape({
  opponent_id: string().uuid().required(),
  opponent_username: string().required(),
});

export const QUEUE_JOINED_SCHEMA = object().shape({
  message: string().required(),
  queue_size: number().required(),
});


// =======================
// UTILS
// =======================

export class WSTestClient {
  private ws: WebSocket | null = null;
  private messageQueue: Message[] = [];
  private resolvers: Array<(msg: Message) => void> = [];

  constructor(private url: string) {}

  async connect(): Promise<WebSocket> {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(this.url);

      this.ws.on('open', () => resolve(this.ws!));
      this.ws.on('error', reject);
      
      this.ws.on('message', (data: Buffer) => {
        const message = JSON.parse(data.toString()) as Message;
        
        if (this.resolvers.length > 0) {
          const resolver = this.resolvers.shift()!;
          resolver(message);
        } else {
          this.messageQueue.push(message);
        }
      });
    });
  }

  async send(message: Message): Promise<void> {
    if (!this.ws) throw new Error('WebSocket not connected');
    
    return new Promise((resolve, reject) => {
      this.ws!.send(JSON.stringify(message), (err) => {
        if (err) reject(err);
        else resolve();
      });
    });
  }

  async receive<T = any>(): Promise<Message<T>> {
    if (this.messageQueue.length > 0) {
      return this.messageQueue.shift()! as Message<T>;
    }

    return new Promise((resolve) => {
      this.resolvers.push(resolve as (msg: Message) => void);
    });
  }

  async close(): Promise<void> {
    if (!this.ws) return;
    
    return new Promise((resolve) => {
      this.ws!.on('close', () => resolve());
      this.ws!.close();
    });
  }

  isOpen(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export function createMessage<T>(type: number, payload: T): Message<T> {
  return { type, payload };
}

export async function waitForMessage<T = any>(
  client: WSTestClient,
  timeout = 5000
): Promise<Message<T>> {
  return Promise.race([
    client.receive<T>(),
    new Promise<Message<T>>((_, reject) =>
      setTimeout(() => reject(new Error('Message timeout')), timeout)
    ),
  ]);
}

/**
 * Validates the structure of the given response data using the specified schema.
 *
 * This function ensures that the response data conforms to the provided schema
 * using Yup validation. It throws an error if the validation fails, indicating
 * any discrepancies in the structure of the data.
 *
 * @param responseData - The data object to be validated.
 * @param {Schema} schema - A Yup schema object that defines the expected structure
 * and constraints for the response data.
 * @returns {Promise<void>} - Resolves if the validation is successful, otherwise throws an error.
 * @throws Will throw an error if the response data does not match the provided schema.
 */
export async function validateResponse(responseData, schema) {
  try {
    // Validar si la estructura sigue el esquema
    await schema.validate(responseData, {
      strict: true,
      abortEarly: false,
    });
    console.log("Record response structure is valid.");
  } catch (error) {
    console.error(
      "Invalid response structure:",
      error.errors,
      "\nReceived\n",
      JSON.stringify(responseData, null, "\t")
    );
    throw new Error("Schema response validation failed.");
  }
}

// --- Axios helpers ---
/**
 * Use when a request that was expected to success fails. If the error is from axios
 * it throws an error with the request response payload, or else it just throws the error as is.
 */
export function handleUnexpectedAxiosError(err: Error) {
  if (axios.isAxiosError(err) && err?.response?.data) {
    throw new Error(
      `Error: status ${err.response.status} \n${JSON.stringify(err.response.data, null, "\t")}`
    );
  } else if (axios.isAxiosError(err) && err?.response?.data === "") {
    throw new Error(`Error: status ${err.response.status}`);
  }
  console.log(err);
  throw err;
}

/**
 * Use when a request is expected to fail. If the error is from axios it runs the `validator` callback,
 * by passing the entire response object or else it just throws the error as is.
 * @param {Error} response Reponse to check
 * @param {Function} validator Callback to check payload, if the error comes from axios.
 * @returns {Error}
 */
export function handleExpectedAxiosError(
  response: object,
  validator: CallableFunction
) {
  if (axios.isAxiosError(response)) {
    validator(response);
  } else {
    throw response;
  }
}