import { HandlerContext, Handlers } from "$fresh/server.ts";

export const handler: Handlers<unknown, AppState> = {
  GET(_req: Request, ctx: HandlerContext<unknown, AppState>) {
    const userID = ctx.state.auth.sub!;
    const todoID = ctx.params.id;
    const todos = ctx.state.store.get(userID);
    const todo = todos.get(todoID);

    return new Response(JSON.stringify(todo));
  },
  async PATCH(req: Request, ctx: HandlerContext<unknown, AppState>): Response {
    const userID = ctx.state.auth.sub!;
    const todoID = ctx.params.id;
    const todos = ctx.state.store.get(userID) || new Map();
    const { checked = false } = await req.json();

    if (todos.has(todoID)) {
      const todo = todos.get(todoID);
      todo.checked = checked;
      todos.set(todoID, todo);
      ctx.state.store.set(userID, todos);
    }

    return new Response(null, { status: 204 });
  },
  DELETE(_req: Request, ctx: HandlerContext<unknown, AppState>): Response {
    const userID = ctx.state.auth.sub!;
    const todoID = ctx.params.id;
    const todos = ctx.state.store.get(userID);

    todos.delete(todoID);
    ctx.state.store.set(userID, todos);

    return new Response(null, { status: 204 });
  },
};
