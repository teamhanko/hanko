import * as jose from 'https://deno.land/x/jose@v4.14.4/index.ts';


declare global {
  type Todo = { todoID: string, description: string, checked: boolean };
  type Store = Map<string, Map<string, Todo>>;

  interface AppState {
    store: Store,
    auth: jose.JWTPayload;
  }
}
