import { HandlerContext, Handlers } from "$fresh/server.ts";

export const handler: Handlers = {
  GET(_req: Request, _ctx: HandlerContext) {
    return new Response("Hello World");
  },
  POST(_req: Request, _ctx: HandlerContext): Response {
    return new Response(null, {status: 200});
  },
};
