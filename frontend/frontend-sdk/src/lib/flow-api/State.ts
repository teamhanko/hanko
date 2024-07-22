import {
  FetchNextState,
  StateName,
  Actions,
  Payloads,
} from "./types/state-handling";
import { Error } from "./types/error";
import { Action } from "./types/action";
import { Input } from "./types/input";

type InputValues<TInput extends Record<string, Input<any>>> = {
  [K in keyof TInput]?: TInput[K]["value"];
};

type CreateAction<TAction extends Action<any>> = (
  inputs: InputValues<TAction["inputs"]>
) => TAction & {
  run: () => Promise<State<any>>;
  validate: () => TAction;
  tryValidate: () => ValidationError | void;
};

type ActionFunctions = {
  [TStateName in keyof Actions]: {
    [TActionName in keyof Actions[TStateName]]: Actions[TStateName][TActionName] extends Action<
      infer Inputs
    >
      ? CreateAction<Action<Inputs>>
      : never;
  };
};

interface StateResponse<TStateName extends StateName> {
  name: StateName;
  status: number;
  payload?: Payloads[TStateName];
  actions?: Actions[TStateName];
  csrf_token: string;
  error: Error;
}

// State class represents a state in the flow
// eslint-disable-next-line require-jsdoc
class State<TStateName extends StateName>
  implements Omit<StateResponse<TStateName>, "actions">
{
  readonly name: StateName;
  readonly payload?: Payloads[TStateName];
  readonly error: Error;
  readonly status: number;
  readonly csrf_token: string;

  readonly #actionDefinitions: Actions[TStateName];
  readonly actions: ActionFunctions[TStateName];

  private readonly fetchNextState: FetchNextState;

  toJSON() {
    return {
      name: this.name,
      payload: this.payload,
      error: this.error,
      status: this.status,
      csrf_token: this.csrf_token,
      actions: this.#actionDefinitions,
    };
  }

  // eslint-disable-next-line require-jsdoc
  constructor(
    { name, payload, error, status, actions, csrf_token }: StateResponse<TStateName>,
    fetchNextState: FetchNextState
  ) {
    this.name = name;
    this.payload = payload;
    this.error = error;
    this.status = status;
    this.csrf_token = csrf_token;
    this.#actionDefinitions = actions;

    // We're doing something really hacky here, but hear me out
    //
    // `actions` is an object like this:
    //
    //     { login_password_recovery: { inputs: { new_password: { min_length: 8, value: "this still needs to be set" } } } }
    //
    // However, we don't want users to have to mutate the `actions` object manually.
    // They WOULD have to do this:
    //
    //     actions.login_password_recovery.inputs.new_password.value = "password";
    //
    // Instead, we're going to wrap the `actions` object in a Proxy.
    // This Proxy transforms the manual mutation you're seeing above into a function call.
    // The following is doing the same thing as the manual mutation above:
    //
    //     actions.login_password_recovery({ new_password: "password" });
    //
    // Okay, there's one difference, the function call creates a copy of the action, so it's not mutating the original object.
    // The newly created action is returned. It also has a `run` method, which sends the action to the server (fetchNextState)
    this.actions = this.#createActionsProxy(actions, csrf_token);

    // Do not remove! `this.fetchNextState` has to be set for `this.#runAction` to work
    this.fetchNextState = fetchNextState;
  }

  /**
   * We get the `actions` object from the server. That object is essentially a definition of actions that can be performed in the current state.
   *
   * For example:
   *
   *     actions = {
   *       login_password_recovery: {
   *         inputs: {
   *           email: { value: undefined, required: true, ... },
   *           password: { value: undefined, required: true, min_length: 8, ... }
   *         }
   *       },
   *       create_account: { inputs: ... },
   *       some_other_action: { inputs: ... },
   *     };
   *
   * The proxy returned by this method creates "action functions".
   *
   * Each action function copies the original definition (`{ inputs: ... }`) and modifies that copy with the inputs provided by the user.
   *
   * In practice, it looks like this:
   *
   *     actions.login_password_recovery({ new_password: "very-secure-123" });
   *     // => { inputs: { password: { value: "very-secure-123", min_length: 8, ... }}}
   *
   * Additionally, helper methods like `run` (to send the action to the server) and `validate` (to validate the inputs; the `inputs` object also contains validation rules)
   */
  #createActionsProxy(actions: Actions[TStateName], csrfToken: string) {
    const runAction = (action: Action<any>) => this.runAction(action, csrfToken);
    const validateAction = (action: Action<any>) => this.validateAction(action);

    return new Proxy(actions, {
      get(target, prop): CreateAction<Action<unknown>> | undefined {
        if (typeof prop === "symbol") return (target as any)[prop];

        type Original = Actions[TStateName][keyof Actions[TStateName]];
        type Prop = keyof typeof target;

        /**
         * This is the action defintion.
         * Running the function returned by this getter creates a **deep copy**
         * with values set by the user.
         */
        const originalAction = target[
          prop as Prop
        ] satisfies Original as Action<unknown>;

        if (originalAction == null) {
          return null;
        }

        return (newInputs: any) => {
          const action = Object.assign(deepCopy(originalAction), {
            validate() {
              validateAction(action);
              return action;
            },
            /**
             * Safe version of `validate` that returns
             */
            tryValidate() {
              try {
                validateAction(action);
              } catch (e) {
                if (e instanceof ValidationError) return e;

                // We still want to throw non-ValidationErrors since they're unexpected (and indicate a bug on our side)
                throw e;
              }
            },
            run() {
              return runAction(action);
            },
          });

          // If `actions` is an object that has inputs,
          //
          // Transform this:
          // actions.login_password_recovery({ new_password: "password" });
          //                                 ^^^^^^^^^^^^^^^^^^^^^^^^^^^^
          // Into this:
          // action.inputs = { new_password: { min_length: 8, value: "password", ... }}
          if (
            action !== null &&
            typeof action === "object" &&
            "inputs" in action
          ) {
            for (const inputName in newInputs) {
              const actionInputs = action.inputs as Record<
                string,
                Input<unknown>
              >;

              if (!actionInputs[inputName]) {
                actionInputs[inputName] = { name: inputName, type: "" };
              }

              actionInputs[inputName].value = newInputs[inputName];
            }
          }

          return action;
        };
      },
    }) satisfies Actions[TStateName] as any;
  }

  runAction(action: Action<any>, csrfToken: string): Promise<State<any>> {
    const data: Record<string, any> = {};

    // Deal with object-type inputs
    // i.e. actions.some_action({ ... })
    //                          ^^^^^^^
    // Other input types would look like this:
    //
    // actions.another_action(1234);
    // actions.yet_another_action("foo");
    //
    // Meaning
    if (
      "inputs" in action &&
      typeof action.inputs === "object" &&
      action.inputs !== null
    ) {
      // This looks horrible, but at this point we're sure that `action.inputs` is a Record<string, Input>
      // Because there are no object-type inputs that AREN'T a Record<string, Input>
      const inputs = action.inputs satisfies object as Record<
        string,
        Input<unknown>
      >;

      for (const inputName in action.inputs) {
        const input = inputs[inputName];

        if (input && "value" in input) {
          data[inputName] = input.value;
        }
      }
    }

    // (Possibly add more input types here?)

    // Use the fetch function to perform the action
    return this.fetchNextState(action.href, {
      input_data: data,
      csrf_token: csrfToken,
    });
  }

  validateAction(action: Action<{ [key: string]: Input<unknown> }>) {
    if (!("inputs" in action)) return;

    for (const inputName in action.inputs) {
      const input = action.inputs[inputName];

      function reject<T>(
        reason: ValidationReason,
        message: string,
        wanted?: T,
        actual?: T
      ) {
        throw new ValidationError({
          reason,
          inputName,
          wanted,
          actual,
          message,
        });
      }

      const value = input.value as any; // TS gets in the way here

      // TODO is !input.value right here? this will also reject empty strings, `0`, ... and will never reject an empty array/object
      if (input.required && !value) {
        reject(ValidationReason.Required, "is required");
      }

      const hasLengthRequirement =
        input.min_length != null || input.max_length != null;

      if (hasLengthRequirement) {
        if (!("length" in value)) {
          reject(
            ValidationReason.InvalidInputDefinition,
            'has min/max length requirement, but is missing "length" property',
            "string",
            typeof value
          );
        }

        if (input.min_length != null && value < input.min_length) {
          reject(
            ValidationReason.MinLength,
            `too short (min ${input.min_length})`,
            input.min_length,
            value.length
          );
        }

        if (input.max_length != null && value > input.max_length) {
          reject(
            ValidationReason.MaxLength,
            `too long (max ${input.max_length})`,
            input.max_length,
            value.length
          );
        }
      }
    }
  }
}

export enum ValidationReason {
  InvalidInputDefinition,
  MinLength,
  MaxLength,
  Required,
}

export class ValidationError<TWanted = undefined> extends Error {
  reason: ValidationReason;
  inputName: string;
  wanted: TWanted;
  actual: TWanted;

  constructor(opts: {
    reason: ValidationReason;
    inputName: string;
    wanted: TWanted;
    actual: TWanted;
    message: string;
  }) {
    super(`"${opts.inputName}" ${opts.message}`);

    this.name = "ValidationError";
    this.reason = opts.reason;
    this.inputName = opts.inputName;
    this.wanted = opts.wanted;
    this.actual = opts.actual;
  }
}

function deepCopy<T>(obj: T): T {
  return JSON.parse(JSON.stringify(obj));
}

export function isState(x: any): x is State<any> {
  return (
    typeof x === "object" &&
    x !== null &&
    "status" in x &&
    "error" in x &&
    "name" in x &&
    Boolean(x.name) &&
    Boolean(x.status)
  );
}

export { State };
