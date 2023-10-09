import { HandlerContext, Handlers } from "$fresh/server.ts";
import { cryptoRandomString } from "https://deno.land/x/crypto_random_string@1.0.0/mod.ts";

export const handler: Handlers<unknown, AppState> = {
  GET(_req: Request, ctx: HandlerContext<unknown, AppState>) {
    const userID: string = ctx.state.auth.sub!;
    const todos = ctx.state.store.get(userID) ?? new Map();
    return new Response(JSON.stringify(Array.from(todos.values())));
  },
  async POST(
    req: Request,
    ctx: HandlerContext<unknown, AppState>,
  ): Promise<Response> {
    const userID = ctx.state.auth.sub!;
    const todoID = cryptoRandomString({ length: 10, type: "alphanumeric" });
    const { description, checked } = await req.json();
    const todos = ctx.state.store.get(userID) || new Map();
    todos.set(todoID, { todoID, description, checked });

    ctx.state.store.set(userID, todos);

    return new Response(JSON.stringify({ todoID }), { status: 201 });
  },
};
