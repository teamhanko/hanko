import { Claims } from "../Dto";

/**
 * Enum-like type defining the actions that can be broadcasted.
 *
 * @ignore
 * @category SDK
 * @subcategory Internal
 */
type Action = "sessionExpired" | "sessionCreated" | "requestLeadership";

/**
 * Interface representing the data structure of a channel event.
 *
 * @interface
 * @property {Action} action - The type of action being broadcasted.
 * @property {Claims=} claims - Optional claims associated with the event.
 * @property {boolean=} is_valid - Optional indication of the session validity.
 * @category SDK
 * @subcategory Internal
 */
export interface BroadcastMessage {
  action: Action;
  claims?: Claims;
  is_valid?: boolean;
}

/**
 * Callback type for handling broadcast messages.
 *
 * @ignore
 */
// eslint-disable-next-line no-unused-vars
type Callback = (msg: BroadcastMessage) => void;

/**
 * Manages inter-tab communication using the BroadcastChannel API.
 *
 * @category SDK
 * @subcategory Internal
 * @param {string} channelName - The name of the broadcast channel.
 * @param {Callback} onSessionExpired - Callback invoked when the session has expired.
 * @param {Callback} onSessionCreated - Callback invoked when a session is created.
 * @param {Callback} onLeadershipRequested - Callback invoked when a leadership request is received.
 */
export class SessionChannel {
  channel: BroadcastChannel; // The broadcast channel used for communication.
  onSessionExpired: Callback; // Callback invoked when the session has expired.
  onSessionCreated: Callback; // Callback invoked when a session is created.
  onLeadershipRequested: Callback; // Callback invoked when a leadership request is received.

  // eslint-disable-next-line require-jsdoc
  constructor(
    channelName: string = "hanko_session",
    onSessionExpired: Callback,
    onSessionCreated: Callback,
    onLeadershipRequested: Callback,
  ) {
    this.onSessionExpired = onSessionExpired;
    this.onSessionCreated = onSessionCreated;
    this.onLeadershipRequested = onLeadershipRequested;

    this.channel = new BroadcastChannel(channelName);
    this.channel.onmessage = this.handleMessage;
  }

  /**
   * Sends a message via the broadcast channel to inform other tabs of session changes.
   *
   * @param {BroadcastMessage} msg - The messsage to broadcast.
   */
  post(msg: BroadcastMessage) {
    this.channel.postMessage(msg);
  }

  /**
   * Handles incoming messages from the broadcast channel.
   *
   * @param {MessageEvent} event - The message event containing the broadcast data.
   * @private
   */
  private handleMessage = (event: MessageEvent) => {
    const data = event.data as BroadcastMessage;
    switch (data.action) {
      case "sessionExpired":
        this.onSessionExpired(data);
        break;
      case "sessionCreated":
        this.onSessionCreated(data);
        break;
      case "requestLeadership":
        this.onLeadershipRequested(data);
        break;
    }
  };
}
