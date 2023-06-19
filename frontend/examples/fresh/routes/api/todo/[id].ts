import { HandlerContext, Handlers } from "$fresh/server.ts";

export const handler: Handlers<unknown, AppState> = {
  GET(req: Request, ctx: HandlerContext<unknown, AppState>) {
    const userID = ctx.state.auth.sub!;
    const todoID = ctx.params.id;
    const todos = ctx.state.store.get(userID);
    const todo = todos.get(todoID);

    return new Response(JSON.stringify(todo));
  },
  async PATCH(req: Request, ctx: HandlerContext<unknown, AppState>): Response {
    const userID = ctx.state.auth.sub!;
    const todoID = ctx.params.id;
    const todos = ctx.state.store.get(userID);
    const data = await req.json();

    if (data.checked) {
      const checked = data.checked;

      if (todos.has(todoID)) {
        todos.get(todoID).checked = checked;
      }
  }
    return new Response(null, {status: 204});
  },
  DELETE(req: Request, ctx: HandlerContext<unknown, AppState>): Response {
    const userID = ctx.state.auth.sub!;
    const todoID = ctx.params.id;
    const todos = ctx.state.store.get(userID);
    todos.delete(todoID);

    return new Response(null, {status: 204});
  },
};
