import {MiddlewareHandlerContext} from "$fresh/server.ts";
import {getCookies} from "$std/http/cookie.ts";
import * as jose from 'https://deno.land/x/jose@v4.14.4/index.ts';

const JWKS_ENDPOINT = `${Deno.env.get("HANKO_API_URL")}/.well-known/jwks.json`;
const store: Store = new Map();

function getToken(req: Request): string | undefined {
  const cookies = getCookies(req.headers);
  const authorization = req.headers.get("authorization")

  if (authorization && authorization.split(" ")[0] === "Bearer")
    return authorization.split(" ")[1]
  else if (cookies.hanko)
    return cookies.hanko
}

export async function handler(req: Request, ctx: MiddlewareHandlerContext<AppState>) {
  const JWKS = jose.createRemoteJWKSet(new URL(JWKS_ENDPOINT), {
    cooldownDuration: 120000,
  });
  const jwt = getToken(req);

  if (!jwt)
    return new Response(null, {status: 401});

  try {
    const {payload} = await jose.jwtVerify(jwt, JWKS);
    ctx.state.auth = payload;
    ctx.state.store = store;

    return await ctx.next();
  } catch (e) {
    console.log(e)
    return new Response(null, {status: 401});
  }
}
