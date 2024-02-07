import {
  FetchNextState,
  StateName,
  Actions,
  Payloads,
} from "./types/state-handling";
import { Error } from "./types/error";
import { Action } from "./types/action";
import { Input } from "./types/input";

type InputValues<I extends Record<string, Input<any>>> = {
  [K in keyof I]: I[K]["value"];
};

type CreateAction<A extends Action<any>> = (
  inputs: InputValues<A["inputs"]>
) => A & { run: () => Promise<State<any>> };

type ActionFunctions = {
  [StateName in keyof Actions]: {
    [ActionName in keyof Actions[StateName]]: Actions[StateName][ActionName] extends Action<
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
  error: Error;
}

// State class represents a state in the flow
// eslint-disable-next-line require-jsdoc
class State<TStateName extends StateName>
  implements Omit<StateResponse<TStateName>, "actions">
{
  readonly name: StateName;
  readonly payload?: Payloads[TStateName];
  readonly actions: ActionFunctions[TStateName];
  readonly error: Error;
  readonly status: number;

  private readonly fetchNextState: FetchNextState;

  // eslint-disable-next-line require-jsdoc
  constructor(
    { name, payload, error, status, actions }: StateResponse<TStateName>,
    fetchNextState: FetchNextState
  ) {
    this.name = name;
    this.payload = payload;
    this.error = error;
    this.status = status;

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
    this.actions = this.#createActionsProxy(actions, fetchNextState);

    this.fetchNextState = fetchNextState;
  }

  #createActionsProxy(
    actions: Actions[TStateName],
    fetchNextState: FetchNextState
  ) {
    return new Proxy(actions, {
      get(target, prop) {
        if (typeof prop === "symbol") return (target as any)[prop];

        type Original = Actions[TStateName][keyof Actions[TStateName]];
        type Prop = keyof typeof target;

        const createAction: CreateAction<Action<unknown>> = (
          newInputs: any
        ) => {
          const action = {
            ...(target[prop as Prop] satisfies Original as Action<unknown>),

            // "run" function that sends the action to the Flow API
            run: () => {
              const data: Record<string, any> = {};

              // Deal with object-type inputs
              // i.e. actions.some_action({ ... })
              //                          ^^^^^^^
              //
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
              return fetchNextState(this.href, {
                input_data: JSON.stringify(data),
              });
            },
          };

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

              actionInputs[inputName].value = newInputs[inputName];
            }
          }

          return action;
        };

        return createAction;
      },
    }) satisfies Actions[TStateName] as any;
  }
}

export { State };
